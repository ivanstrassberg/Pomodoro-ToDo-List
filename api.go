package main

import (
	"database/sql"
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
	if err := json.NewDecoder(r.Body).Decode(createTaskRequest); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "")
	}
	// if flag := isValidRFC3339(createTaskRequest.DueDate); !flag {
	// 	return WriteJSON(w, http.StatusBadRequest, "Bad time format")
	// }
	createdTask := &Task{
		Title:       createTaskRequest.Title,
		Description: createTaskRequest.Description,
		DueDate:     createTaskRequest.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := s.store.CreateTask(createdTask)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	WriteJSON(w, http.StatusOK, createdTask)
	return nil
}
func (s *APIServer) handleGetTasks(w http.ResponseWriter, r *http.Request) error {
	tasksSlice, err := s.store.GetTasks()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "")
		return err
	}
	WriteJSON(w, http.StatusOK, tasksSlice)
	return nil
}

func (s *APIServer) handleGetTaskByID(w http.ResponseWriter, r *http.Request) error {
	idInt, err := getIDFromURL(r)
	if err != nil {
		return err
	}
	taskByID, err := s.store.GetTaskByID(idInt)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusNotFound, "Task not found")
			return err
		}
		WriteJSON(w, http.StatusInternalServerError, "Server error")
		return err
	}
	WriteJSON(w, http.StatusOK, taskByID)
	return nil
}
func (s *APIServer) handleUpdateTaskByID(w http.ResponseWriter, r *http.Request) error {

	idInt, err := getIDFromURL(r)
	createTaskRequest := new(TaskCreateReq)
	if err := json.NewDecoder(r.Body).Decode(createTaskRequest); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "")
	}
	if err != nil {
		return err
	}
	updatedTask, err := s.store.UpdateTask(idInt, *createTaskRequest)
	if err != nil {
		return err
	}
	//something wrong with writejson, just sends empty structs EVERYWHERE!
	//400 404 500 handle
	WriteJSON(w, http.StatusOK, updatedTask)
	return nil
}
func (s *APIServer) handleDeleteTaskByID(w http.ResponseWriter, r *http.Request) error {
	idInt, err := getIDFromURL(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteTask(idInt); err != nil {
		return err
	}

	WriteJSON(w, http.StatusNoContent, "Task deleted")
	// refer to codes
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
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
