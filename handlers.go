package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/polypmer/ghess"
)

// Index page, link to new game
func Index(w http.ResponseWriter,
	r *http.Request) {
	var gameList = make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("games"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// k key v value
			gameList[string(k)] = string(v)
		}
		return nil
	})

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Printf("Error %s Templates", err)
	}

	t.Execute(w, gameList)
}

func NewGame(w http.ResponseWriter,
	r *http.Request) {
	vars := mux.Vars(r)
	color := vars["player"]
	game := ghess.NewBoard()
	if color == "black" {
		game.Move(24, 44)
	}
	// new Board
	// Make first move if black

	// Key Value Pair
	value := []byte(game.Position())
	key := []byte(time.Now().String())
	// Add to Database
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("games"))
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
	// Redirect to View Board
	http.Redirect(w, r, "/view/"+string(key), http.StatusSeeOther)

}

func ViewGame(w http.ResponseWriter,
	r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var pos string
	// Read
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(games)
		if bucket == nil {
			return errors.New("No bucket")
		}
		val := bucket.Get([]byte(id))
		pos = string(val)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	t, err := template.ParseFiles("templates/computer.html")
	if err != nil {
		fmt.Printf("Error %s Templates", err)
	}

	t.Execute(w, pos)
}

// AJAX call to make move
func PlayGame(w http.ResponseWriter,
	r *http.Request) {

}

/* AI Game */
// TODO:
