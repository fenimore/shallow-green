package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/polypmer/ghess"
)

// TODO:
// The keys for games will be unix time.
// Should these get to be a certain time before a certain
// time, then delete them.
// TODO:
// Name ai

var games = []byte("games")

var db *bolt.DB

type Game struct {
	g       ghess.Board
	white   string
	black   string
	created time.Time
}

// Open Bolddb connection

func main() {
	// Handle DB connection
	blt, err := bolt.Open("games.db", 0644, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer blt.Close()
	db = blt

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("games"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// connection
	router := NewRouter()
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
