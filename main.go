package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/polypmer/ghess"
)

var games = []byte("games")

type Game struct {
	g    ghess.Board
	date time.Time
}

// Open Bolddb connection

func main() {
	db, err := bolt.Open("games.db", 0644, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	key := []byte("1")

	// Read
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(games)
		if bucket == nil {
			return errors.New("No bucket")
		}
		val := bucket.Get(key)
		fmt.Println(string(val))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Write Put

	value := []byte("hello")
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(games)
		if err != nil {
			return err
		}
		err = bucket.Put(key, value)
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
