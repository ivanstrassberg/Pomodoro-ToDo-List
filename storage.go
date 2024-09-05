package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateTask(*Task) (Task, error)
	GetTasks() ([]TaskID, error)
}
type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) Init() error {
	_, err := s.db.Query(`
	create table if not exists task (
		id serial primary key,
		title varchar(255) not null,
		description text,
		due_date timestamp with time zone,
		created_at timestamp default current_timestamp,
		updated_at timestamp default current_timestamp
		);
`)
	if err != nil {
		return err
	}
	fmt.Println("DB initialized")
	return nil

}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres port=5433 dbname=todo_list password=root sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {

		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) CreateTask(t *Task) (Task, error) {

	queryStr := `insert into task (title, description, due_date, created_at, updated_at) values ($1,$2,$3,$4,$5) `
	rows, err := s.db.Query(queryStr, t.Title, t.Description, t.DueDate, t.ApplyCurrentTimeToTask("CreatedAt"), t.ApplyCurrentTimeToTask("UpdatedAt"))
	if err != nil {
		return Task{}, err
	}
	task := new(Task)
	if err := ScanIntoStruct(rows, &task); err != nil {
		return Task{}, err
	}
	return *task, nil
}

func (s *PostgresStore) GetTasks() ([]TaskID, error) {

	var taskSlice []TaskID
	query := `select * from task`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		task, err := scanIntoTask(rows)
		if err != nil {
			return nil, err
		}
		taskSlice = append(taskSlice, task)
	}

	return taskSlice, nil
}

func scanIntoTask(rows *sql.Rows) (TaskID, error) {
	var task TaskID
	err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return TaskID{}, err
	}
	return task, nil
}

func ScanIntoRow(r *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("dest must point to struct to scan into")
	}
	v = v.Elem()
	numElements := v.NumField()
	fields := make([]interface{}, numElements)
	for i := 0; i < numElements; i++ {
		fields[i] = v.Field(i).Addr().Interface()
	}
	if err := r.Scan(fields...); err != nil {
		return err
	}
	return nil
}

func ScanIntoStruct(r *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return errors.New("dest must be a pointer to a slice")
	}
	elemType := v.Elem().Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("dest slice must contain struct elements")
	}
	sliceValue := v.Elem()
	for r.Next() {
		elemPtr := reflect.New(elemType).Interface()

		if err := ScanIntoRow(r, elemPtr); err != nil {
			return err
		}

		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(elemPtr).Elem()))
	}

	return r.Err()
}
