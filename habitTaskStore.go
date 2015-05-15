package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"bytes"

	"github.com/Luzifer/habitscheduler/habitrpg"

	"github.com/robfig/cron"
	"github.com/satori/go.uuid"
	"github.com/xuyu/goredis"
)

type HabitTaskStore struct {
	Tasks []HabitTask `json:",omitempty"`

	config          *Config        `json:"-"`
	redisConnection *goredis.Redis `json:"-"`
}

func NewHabitTaskStore(config *Config, redisConnection *goredis.Redis) *HabitTaskStore {
	return &HabitTaskStore{
		config:          config,
		redisConnection: redisConnection,
		Tasks:           []HabitTask{},
	}
}

func (h *HabitTaskStore) Save() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}

	err = h.redisConnection.Set(h.config.RedisStoreKey, string(data), 0, 0, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *HabitTaskStore) Load() error {
	data, err := h.redisConnection.Get(h.config.RedisStoreKey)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		data = []byte("{}")
	}

	err = json.Unmarshal(data, h)
	if err != nil {
		return err
	}

	return nil
}

func (h *HabitTaskStore) doHTTPRequest(method, contentType, urlStr string, body io.Reader, targetVar interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("https://habitrpg.com:443/api/v2%s", urlStr), body)
	if err != nil {
		return err
	}
	req.Header.Add("x-api-key", h.config.HabitRPGAPIToken)
	req.Header.Add("x-api-user", h.config.HabitRPGUserID)
	req.Header.Add("Content-Type", contentType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(responseBody, targetVar)
	if err != nil {
		return err
	}

	return nil
}

func (h *HabitTaskStore) UpdateStates() {
	res := []habitrpg.Task{}
	err := h.doHTTPRequest("GET", "application/json", "/user/tasks", nil, &res)
	if err != nil {
		panic(err)
	}

	for i, _ := range h.Tasks {
		task := &h.Tasks[i]
		if len(task.LastTaskID) < 1 {
			continue
		}

		taskFound := false
		for _, htask := range res {
			if htask.ID == task.LastTaskID {
				task.IsCompleted = htask.Completed
				task.updateNextEntryTime(htask.DateCompleted)
				if task.IsCompleted {
					task.LastTaskID = ""
				}
				taskFound = true
			}
		}
		if !taskFound {
			task.IsCompleted = true
			task.updateNextEntryTime(time.Now())
			task.LastTaskID = ""
		}
	}
}

func (h *HabitTaskStore) CreateDueTasks() {
	log.Println("Creating tasks...")
	for i, _ := range h.Tasks {
		task := &h.Tasks[i]
		if task.IsCompleted && time.Now().After(task.NextEntryDate) {
			newTask := habitrpg.Task{
				Type:        "todo",
				Text:        task.Title,
				DateCreated: time.Now(),
			}

			data, err := json.Marshal(newTask)
			if err != nil {
				panic(err)
			}

			reader := bytes.NewReader(data)
			res := habitrpg.Task{}
			err = h.doHTTPRequest("POST", "application/json", "/user/tasks", reader, &res)
			if err != nil {
				panic(err)
			}

			task.LastTaskID = res.ID
			task.IsCompleted = false
		}
	}
}

type HabitTask struct {
	ID string

	Title         string
	LastTaskID    string
	NextEntryDate time.Time
	IsCompleted   bool

	RepeatHours     int
	RepeatCron      bool
	RepeatCronEntry string
}

func NewCronTask(title, cronSpec string) (*HabitTask, error) {
	scheduler, err := cron.Parse(cronSpec)
	if err != nil {
		return nil, err
	}

	return &HabitTask{
		ID:              uuid.NewV4().String(),
		Title:           title,
		RepeatCron:      true,
		RepeatCronEntry: cronSpec,
		NextEntryDate:   scheduler.Next(time.Now()),
		IsCompleted:     true,
	}, nil
}

func NewHourRepeatTask(title string, repeatAfter int) *HabitTask {
	return &HabitTask{
		ID:            uuid.NewV4().String(),
		Title:         title,
		RepeatHours:   repeatAfter,
		RepeatCron:    false,
		NextEntryDate: time.Now(),
		IsCompleted:   true,
	}
}

func (t *HabitTask) updateNextEntryTime(dateCompleted time.Time) {
	if t.RepeatCron {
		scheduler, _ := cron.Parse(t.RepeatCronEntry)
		t.NextEntryDate = scheduler.Next(dateCompleted)
	} else {
		t.NextEntryDate = dateCompleted.Add(time.Duration(t.RepeatHours) * time.Hour)
	}
}
