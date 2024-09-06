package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{listenAddr: listenAddr, store: store}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/home", makeHTTPHandleFunc(s.handleHome))
	mux.HandleFunc("POST /tasks", makeHTTPHandleFunc(s.handlePostTasks))
	mux.HandleFunc("GET /tasks", makeHTTPHandleFunc(s.handleGetTasks))
	mux.HandleFunc("GET /tasks/{id}", makeHTTPHandleFunc(s.handleGetTaskByID))
	mux.HandleFunc("PUT /tasks/{id}", makeHTTPHandleFunc(s.handleUpdateTaskByID))
	mux.HandleFunc("DELETE /tasks/{id}", makeHTTPHandleFunc(s.handleDeleteTaskByID))
	//
	log.Println("JSON API server running on port", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, mux); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

func (s *APIServer) handleHome(w http.ResponseWriter, r *http.Request) error {
	WriteJSON(w, http.StatusOK, "all good")
	return nil
}

func (s *APIServer) handlePostTasks(w http.ResponseWriter, r *http.Request) error {
	createTaskRequest := new(TaskCreateReq)
	jsonDecoder := json.NewDecoder(r.Body)
	// jsonDecoder.DisallowUnknownFields()
	if err := jsonDecoder.Decode(createTaskRequest); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "wrong data format")
	}
	// if flag := isValidRFC3339(createTaskRequest.DueDate); !flag {
	// 	return WriteJSON(w, http.StatusBadRequest, "Bad time format")
	// }
	parsed, err := time.Parse(time.RFC3339, createTaskRequest.DueDate)
	createdTask := &Task{
		Title:       createTaskRequest.Title,
		Description: createTaskRequest.Description,
		DueDate:     parsed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	retTaskID, err := s.store.CreateTask(createdTask)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, "server issue")
	}

	WriteJSON(w, http.StatusOK, TaskID{
		ID: retTaskID.ID,
		Task: Task{
			Title:       createdTask.Title,
			Description: createTaskRequest.Description,
			DueDate:     parsed,
			CreatedAt:   createdTask.CreatedAt,
			UpdatedAt:   createdTask.UpdatedAt,
		},
	})
	return nil
}
func (s *APIServer) handleGetTasks(w http.ResponseWriter, r *http.Request) error {
	tasksSlice, err := s.store.GetTasks()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "server issue")
		return err
	}
	WriteJSON(w, http.StatusOK, tasksSlice)
	return nil
}

func (s *APIServer) handleGetTaskByID(w http.ResponseWriter, r *http.Request) error {
	idInt, err := getIDFromURL(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "invalid ID format")
	}
	taskByID, err := s.store.GetTaskByID(idInt)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Server issue")
		return err
	}
	if taskByID == nil {
		return WriteJSON(w, http.StatusNotFound, "Task not found")
	}
	return WriteJSON(w, http.StatusOK, taskByID)
}
func (s *APIServer) handleUpdateTaskByID(w http.ResponseWriter, r *http.Request) error {

	idInt, err := getIDFromURL(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "invalid ID format")
	}
	createTaskRequest := new(TaskCreateReq)
	if err := json.NewDecoder(r.Body).Decode(createTaskRequest); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "wrong data format")
	}
	if err != nil {
		return err
	}
	updatedTask, err, rowsAffected := s.store.UpdateTask(idInt, *createTaskRequest)
	if updatedTask == nil && rowsAffected >= 0 {
		return WriteJSON(w, http.StatusNotFound, "task not found")
	}
	if err != nil {
		fmt.Println(err)
		return WriteJSON(w, http.StatusInternalServerError, "server issue")
	}

	WriteJSON(w, http.StatusOK, updatedTask)
	return nil
}
func (s *APIServer) handleDeleteTaskByID(w http.ResponseWriter, r *http.Request) error {
	idInt, err := getIDFromURL(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "invalid ID format")
	}
	err, rowsAffected := s.store.DeleteTask(idInt)

	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, "server issue")
	}
	if rowsAffected == 0 {
		return WriteJSON(w, http.StatusNotFound, "task not found")
	}
	if rowsAffected == -1 {
		return WriteJSON(w, http.StatusInternalServerError, "server issue")
	}

	w.WriteHeader(http.StatusNoContent)
	return nil

}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type ApiError struct {
	Error string `json:"error"`
}
type APIfunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f APIfunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}
func isValidRFC3339(dateStr string) bool {
	_, err := time.Parse(time.RFC3339, dateStr)
	return err == nil
}

func getIDFromURL(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("invalid ID format")
	}
	return idInt, nil
}
