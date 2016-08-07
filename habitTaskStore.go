package main

import (
	"encoding/json"
	"fmt"
	"io"
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

	redisConnection *goredis.Redis `json:"-"`
}

func NewHabitTaskStore(redisConnection *goredis.Redis) *HabitTaskStore {
	return &HabitTaskStore{
		redisConnection: redisConnection,
		Tasks:           []HabitTask{},
	}
}

func (h *HabitTaskStore) Save() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}

	err = h.redisConnection.Set(config.RedisStoreKey, string(data), 0, 0, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *HabitTaskStore) Load() error {
	data, err := h.redisConnection.Get(config.RedisStoreKey)
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
	req, err := http.NewRequest(method, fmt.Sprintf("https://habitrpg.com:443/api/v3%s", urlStr), body)
	if err != nil {
		return err
	}
	req.Header.Add("x-api-key", config.HabitRPGAPIToken)
	req.Header.Add("x-api-user", config.HabitRPGUserID)
	req.Header.Add("Content-Type", contentType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("Unexpected status code received: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(targetVar); err != nil {
		return err
	}

	return nil
}

func (h *HabitTaskStore) UpdateStates() error {
	res := struct {
		Data []habitrpg.Task `json:"data"`
	}{}

	if err := h.doHTTPRequest("GET", "application/json", "/tasks/user", nil, &res); err != nil {
		return fmt.Errorf("Unable to fetch current tasks: %s", err)
	}

	for i, _ := range h.Tasks {
		task := &h.Tasks[i]
		if task.LastTaskID == "" {
			if !task.IsCompleted {
				task.IsCompleted = true
			}
			continue
		}

		taskFound := false
		for _, htask := range res.Data {
			if htask.ID == task.LastTaskID {
				task.IsCompleted = htask.Completed
				task.updateNextEntryTime(htask.DateCompleted, false)
				if task.IsCompleted {
					task.LastTaskID = ""
				}
				taskFound = true
			}
		}
		if !taskFound {
			task.IsCompleted = true
			task.updateNextEntryTime(time.Now(), false)
			task.LastTaskID = ""
		}
	}
	return nil
}

func (h *HabitTaskStore) CreateDueTasks() error {
	log.Println("Creating tasks...")
	for i, _ := range h.Tasks {
		task := &h.Tasks[i]
		if task.IsCompleted && time.Now().After(task.NextEntryDate) {
			newTask := habitrpg.Task{
				Type:        "todo",
				Text:        task.Title,
				DateCreated: time.Now(),
			}

			buf := bytes.NewBuffer([]byte{})
			if err := json.NewEncoder(buf).Encode(newTask); err != nil {
				return fmt.Errorf("Unable to encode new task: %s", err)
			}

			res := habitrpg.Task{}
			if err := h.doHTTPRequest("POST", "application/json", "/tasks/user", buf, &res); err != nil {
				return fmt.Errorf("Unable to create new task with API: %s", err)
			}

			task.LastTaskID = res.ID
			task.IsCompleted = false
		}
	}
	return nil
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

func NewTaskWithChecks(input []byte) (*HabitTask, error) {
	tmp := HabitTask{}
	err := json.Unmarshal(input, &tmp)
	if err != nil {
		return nil, fmt.Errorf("Could not deserialize JSON: %s", err)
	}

	out := &HabitTask{
		ID:          uuid.NewV4().String(),
		Title:       tmp.Title,
		IsCompleted: true,
		RepeatHours: tmp.RepeatHours,
		RepeatCron:  false,
	}

	if tmp.RepeatCron && len(tmp.RepeatCronEntry) > 0 {
		_, err := cron.Parse(tmp.RepeatCronEntry)
		if err != nil {
			return nil, fmt.Errorf("Could not parse cron format: %s", err)
		}

		out.RepeatCron = true
		out.RepeatCronEntry = tmp.RepeatCronEntry
	}

	if out.RepeatHours == 0 && out.RepeatCron == false {
		return nil, fmt.Errorf("You must specify at least one of RepeatHours or RepeatCronEntry")
	}

	out.updateNextEntryTime(time.Now(), true)

	return out, nil
}

func (t *HabitTask) updateNextEntryTime(dateCompleted time.Time, initial bool) {
	if initial {
		t.NextEntryDate = time.Now()
	} else {
		t.NextEntryDate = dateCompleted.Add(time.Duration(t.RepeatHours) * time.Hour)
	}

	if t.RepeatCron {
		scheduler, _ := cron.Parse(t.RepeatCronEntry)
		t.NextEntryDate = scheduler.Next(t.NextEntryDate)
	}
}
