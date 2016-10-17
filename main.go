package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/polypmer/ghess"
)

type Game struct {
	g    ghess.Board
	date time.Time
}

// Open Bolddb connection

func main() {
	// connection
	router := NewRouter()
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
