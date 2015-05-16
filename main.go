package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"github.com/xuyu/goredis"
)

var (
	config          *Config
	redisConnection *goredis.Redis
	habitRPG        *HabitTaskStore
)

func init() {
	var err error

	config = LoadConfig()

	redisConnection, err = goredis.DialURL(config.RedisAddress)
	if err != nil {
		log.Printf("Error while connecting to Redis: %s", err)
		os.Exit(1)
	}

	habitRPG = NewHabitTaskStore(config, redisConnection)
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
		}
	})
	c.AddFunc(config.CronCreateTask, habitRPG.CreateDueTasks)
	c.AddFunc(config.CronUpdateTasks, habitRPG.UpdateStates)
	c.Start()

	// API interface
	r := mux.NewRouter()

	v1 := r.PathPrefix("/v1/").Subrouter()
	v1.HandleFunc("/tasks", handleCreateTask).Methods("POST")
	v1.HandleFunc("/tasks", handleGetTasks).Methods("GET")
	v1.HandleFunc("/tasks/{taskid}", handleDeleteTask).Methods("DELETE")

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
