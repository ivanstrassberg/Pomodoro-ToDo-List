package main

import "fmt"

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		fmt.Println(err)
	}
	err = store.Init()
	if err != nil {
		fmt.Println(err)
	}

	server := NewAPIServer(":8888", store)

	server.Run()

}
