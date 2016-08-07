package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"github.com/xuyu/goredis"
)

var (
	config struct {
		RedisAddress  string `flag:"redis-url" default:"" description:"Connectionstring to redis server"`
		RedisStoreKey string `flag:"redis-key" default:"habitrpg-tasks" description:"Key to store the data in"`

		ListenAddress string `flag:"listen" default:":3000" description:"Address incl. port to have the API listen on"`

		HabitRPGUserID   string `flag:"habit-user" default:"" description:"User-ID from API page in HabitRPG"`
		HabitRPGAPIToken string `flag:"habit-token" default:"" description:"API-Token for that HabitRPG user"`

		CronCreateTask  string `flag:"cron-create" default:"0 * * * * *" description:"Cron entry for creating new tasks"`
		CronSaveToRedis string `flag:"cron-persist" default:"0 * * * * *" description:"Cron entry for saving data to Redis"`
		CronUpdateTasks string `flag:"cron-update" default:"10 */5 * * * *" description:"Cron entry for fetchin task updates from HabitRPG"`
	}
	redisConnection *goredis.Redis
	habitRPG        *HabitTaskStore
)

func init() {
	var err error
	rconfig.Parse(&config)

	redisConnection, err = goredis.DialURL(config.RedisAddress)
	if err != nil {
		log.Printf("Error while connecting to Redis: %s", err)
		os.Exit(1)
	}

	habitRPG = NewHabitTaskStore(redisConnection)
	err = habitRPG.Load()
	if err != nil {
		log.Printf("Error while loading HabitRPG store: %s", err)
		os.Exit(1)
	}
}

func main() {
	c := cron.New()
	c.AddFunc(config.CronSaveToRedis, func() {
		err := habitRPG.Save()
		if err == nil {
			log.Println("Save to Redis: Success")
		} else {
			log.Printf("Save to Redis: %s\n", err)
			os.Exit(1)
		}
	})
	c.AddFunc(config.CronCreateTask, func() {
		if err := habitRPG.CreateDueTasks(); err != nil {
			log.Printf("An error ocurred while creating tasks: %s", err)
		}
	})
	c.AddFunc(config.CronUpdateTasks, func() {
		if err := habitRPG.UpdateStates(); err != nil {
			log.Printf("An error ocurred while fetching tasks: %s", err)
		}
	})
	c.Start()

	// API interface
	r := mux.NewRouter()

	v1 := r.PathPrefix("/v1/").Subrouter()
	v1.HandleFunc("/tasks", handleCreateTask).Methods("POST")
	v1.HandleFunc("/tasks", handleGetTasks).Methods("GET")
	v1.HandleFunc("/tasks/{taskid}", handleDeleteTask).Methods("DELETE")
	v1.HandleFunc("/tasks/{taskid}/trigger", handleTaskTrigger).Methods("POST")

	http.Handle("/", &MyServer{r})
	http.ListenAndServe(config.ListenAddress, nil)
}

func handleCreateTask(res http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(res, "Unable to read task data", http.StatusInternalServerError)
		return
	}
	task, err := NewTaskWithChecks(body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	habitRPG.Tasks = append(habitRPG.Tasks, *task)
	habitRPG.Save()

	res.Header().Add("Content-Type", "text/plain")
	res.Write([]byte("OK"))
}

func handleGetTasks(res http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(habitRPG.Tasks)

	res.Header().Add("Content-Type", "application/json")
	res.Write(data)
}

func handleDeleteTask(res http.ResponseWriter, r *http.Request) {
	tmp := []HabitTask{}
	vars := mux.Vars(r)

	for _, task := range habitRPG.Tasks {
		if task.ID != vars["taskid"] {
			tmp = append(tmp, task)
		}
	}

	habitRPG.Tasks = tmp

	res.Header().Add("Content-Type", "text/plain")
	res.Write([]byte("OK"))
}

func handleTaskTrigger(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res.Header().Add("Content-Type", "text/plain")

	for i := range habitRPG.Tasks {
		if habitRPG.Tasks[i].ID == vars["taskid"] {
			habitRPG.Tasks[i].NextEntryDate = time.Now()
			http.Error(res, "OK", http.StatusOK)
			return
		}
	}

	res.Write([]byte("Not Found"))
	http.Error(res, "Not found", http.StatusNotFound)
}

type MyServer struct {
	r *mux.Router
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(rw, req)
}
