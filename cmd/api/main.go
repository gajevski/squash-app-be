package main

import (
	"fmt"
	"squash-app-be/internal/auth"
	"squash-app-be/internal/server"
)

func main() {

	auth.NewAuth()
	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
