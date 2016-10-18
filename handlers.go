package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
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

func About(w http.ResponseWriter,
	r *http.Request) {
	t, err := template.ParseFiles("templates/about.html")
	if err != nil {
		fmt.Printf("Error %s Templates", err)
	}
	t.Execute(w, nil)
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
	key := []byte(time.Now().Format("15:04:05"))
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

type Game struct {
	Position string
	Id       string
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
	g := Game{Position: pos, Id: id}

	t.Execute(w, g)
}

// Move is what the browser needs to make calls etc
type Move struct {
	Position string `json:"position"`
	Message  string `json:"message"`
	LastMove string `json:"target"`
	LastOrig string `json:"origin"`
	GameId   string `json:"id"`
}

// AJAX call to make move
func PlayGame(w http.ResponseWriter,
	r *http.Request) {
	// Passed Parameters
	vars := mux.Vars(r)
	id := vars["id"]
	orig := vars["orig"]
	dest := vars["dest"]
	var pos string
	// Get game from DB
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
	// Set up board
	game := ghess.NewBoard()
	err = game.LoadFen(pos)
	if err != nil {
		fmt.Println(err)
	}
	// Make move and ask AI
	mv := &Move{}
	err = game.ParseStand(orig, dest)
	if err != nil {
		mv = &Move{
			Position: game.Position(),
			Message:  "> That's not a Valid Move!",
			GameId:   id,
		}
	} else {
		if game.Checkmate {
			msg := "> I've been Checkmated!"
			mv = &Move{
				Position: game.Position(),
				Message:  msg,
				GameId:   id,
			}
		} else {
			now := time.Now()
			state, err := ghess.MiniMaxPruning(0, 3, ghess.GetState(&game))
			if err != nil {
				fmt.Println("> Minimax broken")
			}
			game.Move(state.Init[0], state.Init[1])
			msg := fmt.Sprintf("> Your Move, <i>my move took %s</i>",
				time.Since(now))
			if game.Checkmate {
				msg = "> Game Over, Checkmate"
			} else if game.Check {
				msg = fmt.Sprintf("> Check! My move took %s", time.Since(now))
			}
			mv = &Move{
				Position: game.Position(),
				Message:  msg,
				LastMove: game.PieceMap[state.Init[1]],
				LastOrig: game.PieceMap[state.Init[0]],
				GameId:   id,
			}

		}
		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(games)
			err = bucket.Put([]byte(id), []byte(game.Position()))
			if err != nil {
				return err
			}
			return nil
		})
	}
	js, err := json.Marshal(mv)
	if err != nil {
		fmt.Println(err)
	}
	w.Write([]byte(js))

}

/* Websockets! */

func NewChallenge(w http.ResponseWriter,
	r *http.Request) {
	game := ghess.NewBoard()

	// Key Value Pair
	value := []byte(game.Position())
	key := []byte(time.Now().Format("15:04:05"))
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
	http.Redirect(w, r, "/challenge/"+string(key), http.StatusSeeOther)
}

func ViewChallenge(w http.ResponseWriter,
	r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var pos string
	// Template
	t, err := template.ParseFiles("templates/versus.html")
	if err != nil {
		fmt.Printf("Error %s Templates", err)
	}
	// Get From Database
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("games"))
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

	g := Game{Position: pos, Id: id}

	t.Execute(w, g)
}

func WebSocket(w http.ResponseWriter,
	r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println("Web Socket?", hub)
	serveWs(id, w, r)
}

// serveWs handles websocket requests from the peer.
func serveWs(id string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Web Socket>>>", hub)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	fmt.Println(client)
	fmt.Println(client.hub)
	fmt.Println(client.hub.register)
	client.hub.register <- client

	var pos string
	// Get game from DB
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("games"))
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
	// Set up board
	game := ghess.NewBoard()
	err = game.LoadFen(pos)
	if err != nil {
		fmt.Println(err)
	}

	go client.writePump(game)
	// So every time this handler is called
	// the client reads the pump
	client.readPump()
}
