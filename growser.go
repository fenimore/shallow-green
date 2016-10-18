package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

// TODO:
// The keys for games will be unix time.
// Should these get to be a certain time before a certain
// time, then delete them.
// TODO:
// Name ai

var games = []byte("games")

var db *bolt.DB
var hub *Hub

// Open Bolddb connection

func main() {
	portFlag := flag.String("port", "8080", "the server port, prefixed by :")
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

	// Launch websocket hub
	hub := newHub()
	fmt.Println(hub)
	go hub.run()

	// connection
	router := NewRouter()
	//http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("static/img"))))
	fmt.Println("Serving Chess on :" + *portFlag)
	//	err = http.ListenAndServe(":"+os.Getenv("PORT"), router) //HEROKU
	err = http.ListenAndServe(":"+*portFlag, router)
	if err != nil {
		fmt.Println(err)
	}
}
