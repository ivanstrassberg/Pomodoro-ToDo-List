package main

import (
	"reflect"
	"time"
)

type TaskCreateReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type TaskID struct {
	ID int `json:"id"`
	Task
}

func (t *Task) ApplyCurrentTimeToTask(params ...string) string {
	currentTime := time.Now().Format(time.RFC3339)
	v := reflect.ValueOf(t).Elem()

	for _, field := range params {
		f := v.FieldByName(field)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			f.SetString(currentTime)
		}
	}
	return currentTime
}
