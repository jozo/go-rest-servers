package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jozo/go-rest-servers/store"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"
)

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	server := NewTaskServer()

	router.HandleFunc("/task/", server.createTaskHandler).Methods("POST")
	router.HandleFunc("/task/", server.getAllTasksHandler).Methods("GET")
	router.HandleFunc("/task/{id:[0-9]+}/", server.getTaskHandler).Methods("GET")

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}

type taskServer struct {
	store *store.TaskStore
}

func NewTaskServer() *taskServer {
	store := store.New()
	return &taskServer{store: store}
}

func (ts *taskServer) getAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all tasks at %s\n", req.URL.Path)

	allTasks := ts.store.GetAllTasks()
	renderJSON(w, allTasks)
}

func (ts *taskServer) createTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling task create at %s\n", req.URL.Path)

	type RequestTask struct {
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}

	type ResponseId struct {
		Id int `json:"id"`
	}

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var rt RequestTask
	if err := dec.Decode(&rt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := ts.store.CreateTask(rt.Text, rt.Tags, rt.Due)
	renderJSON(w, ResponseId{Id: id})
}

func (ts *taskServer) getTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get task at %s\n", req.URL.Path)

	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	task, err := ts.store.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	renderJSON(w, task)
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
