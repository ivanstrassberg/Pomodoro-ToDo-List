Инструкция по запуску.
1 Создать папку, в нее скачать проект с помощью  
"git pull https://github.com/ivanstrassberg/todo_list-Go-API.git"
2 Выполнить команду "go mod tidy"
3 В строке 43 storage.go изменить параметры на ту БД, на которую удобно, либо создать новую с приведенным именем и параметрами: connStr := "user=postgres port=5433 dbname=todo_list password=root sslmode=disable"
4 Выполнить "go run ./"
5 Все должно работать!
Ниже пример тела json запроса
{
  "title": "test title",
  "description": "test description",
  "due_date": "2024-09-06T12:51:34.548908+07:00"
}
Результат работы можно проверить через curl: "curl -X POST http://localhost:8888/tasks -H "Content-Type: application/json" -d '{
  "title": "test title",
  "description": "test description",
  "due_date": "2024-09-06T12:51:34.548908+07:00"
}'

Удобно проверить через расширение Thunder Client
![image](https://github.com/user-attachments/assets/86612bfd-4156-41f5-9134-059d3c7562cd)
URL: http://localhost:8888/tasks
![image](https://github.com/user-attachments/assets/4a75cc66-650a-4b01-87dc-fd0b264eaf5d)
Выбор метода слева, тело запроса в графе Body

