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
	GetTaskByID(int) (TaskID, error)
	UpdateTask(int, TaskCreateReq) (*TaskID, error)
	DeleteTask(int) error
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
		taskSlice = append(taskSlice, *task)
	}
	defer rows.Close()
	return taskSlice, nil
}

// todo handle the cases of failure
func (s *PostgresStore) GetTaskByID(id int) (TaskID, error) {

	var taskRet TaskID
	query := fmt.Sprintf("select * from task where id = %v", id)
	rows, err := s.db.Query(query)

	if err != nil {
		if err == sql.ErrNoRows {
			return TaskID{}, err
		}
		return TaskID{}, err
	}
	for rows.Next() {
		task, err := scanIntoTask(rows)
		if err != nil {
			return TaskID{}, err
		}
		taskRet = *task
	}
	defer rows.Close()

	return taskRet, nil
}

func (s *PostgresStore) UpdateTask(id int, task TaskCreateReq) (*TaskID, error) {
	query := `update task
	set title = $1, description = $2, due_date = $3, updated_at = NOW()
	where id = $4`
	rows, err := s.db.Query(query, task.Title, task.Description, task.DueDate, id)
	if err != nil {
		return nil, err
	}
	var taskRet TaskID
	for rows.Next() {
		task, err := scanIntoTask(rows)
		if err != nil {
			return nil, err
		}
		taskRet = *task
	}
	return &taskRet, nil
}

func (s *PostgresStore) DeleteTask(id int) error {
	query := `delete from task where id = $1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func scanIntoTask(rows *sql.Rows) (*TaskID, error) {
	var task TaskID
	err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &task, nil
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
