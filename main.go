package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
	c.AddFunc("0 * * * * *", func() {
		err := habitRPG.Save()
		if err == nil {
			log.Println("Save to Redis: Success")
		} else {
			log.Printf("Save to Redis: %s\n", err)
		}
	})
	c.AddFunc("0 * * * * *", habitRPG.CreateDueTasks)
	c.AddFunc("0 */5 * * * *", habitRPG.UpdateStates)
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
	var task *HabitTask
	var err error
	switch r.FormValue("taskType") {
	case "cron":
		task, err = NewCronTask(r.FormValue("title"), r.FormValue("cron"))
		if err != nil {
			http.Error(res, fmt.Sprintf("Unable to create task: %s", err), http.StatusInternalServerError)
			return
		}
	case "repeat":
		rep, err := strconv.Atoi(r.FormValue("repeat"))
		if err != nil {
			http.Error(res, "Parameter 'repeat' has to be int.", http.StatusInternalServerError)
			return
		}
		task = NewHourRepeatTask(r.FormValue("title"), rep)
	default:
		http.Error(res, "Please specify task data.", http.StatusInternalServerError)
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
