package main

import (
	"fmt"
	//"time"

	//"github.com/yamazakk1/go-pastebin/internal/app"
	"github.com/yamazakk1/go-pastebin/server"
)

func main() {
	server := server.NewServer()
	err := server.Start(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
