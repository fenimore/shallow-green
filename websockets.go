package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"github.com/polypmer/ghess"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
Websocket structs and functions?!?
*/

/* Hub Functions */

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
// Hub struct for websockets
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// newHub returns a pointer to a new Hub
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// run starts a Hub select switch for
// accepting and disconnecting clients
// and passing on incoming messages.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// This is the json passed from
// the javascript websockets front end
// It's type dictates what kind of broadcast
// I'll be doing
type inCome struct {
	Type        string `json:"type"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Message     string `json:"message"`
	Id          string `json:"id"`
}

// outGo struct sends data back to
// chessboardjs with error position
// and message depending on kind of
// broadcast.
type outGo struct {
	Type     string
	Position string
	Message  string
	Error    string
}

/* Client Functions */

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// Hint: just cause
	c.conn.SetReadLimit(maxMessageSize)              // Why?
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) // why?
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// readPump, unlike write, is not a goroutine
	// And each client runs an infinite loop,
	// Until the connection closes, or there is an error
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// write writes a message with the given message type and payload.
func (c *Client) write(mt int, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump(g ghess.Board) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	var feedback string // For sending info to client
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// read json from message
			msg := inCome{}
			json.Unmarshal([]byte(message), &msg)
			switch msg.Type {
			case "move":
				mv := &outGo{}
				err = g.ParseStand(msg.Origin,
					msg.Destination)
				fen := g.Position()
				info := g.Stats()
				check, _ := strconv.ParseBool(info["check"])
				checkmate, _ := strconv.ParseBool(
					info["checkmate"])
				if check {
					feedback = "Check!"
				} else if checkmate {
					feedback = "Checkmate!"
				}
				if err != nil {
					mv = &outGo{
						Type:     "move",
						Position: fen,
						Error:    err.Error(),
					}
				} else {
					mv = &outGo{
						Type:     "move",
						Position: fen,
						Error:    feedback,
					}
				}
				feedback = ""
				// Marshal into json response
				j, _ := json.Marshal(mv)
				// Update the DB
				err := db.Update(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte("challenges"))
					err = bucket.Put([]byte(msg.Id), []byte(fen))
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					fmt.Println(err)
				}
				// Write Message to Clien
				w.Write([]byte(j))
			case "message":
				chat := &outGo{
					Type:    "message",
					Message: msg.Message,
				}
				j, _ := json.Marshal(chat)
				w.Write([]byte(j))
			case "connection":
				// Should this be put elsewhere?
				chat := &outGo{
					Type:    "connection",
					Message: msg.Message,
				}
				j, _ := json.Marshal(chat)
				w.Write([]byte(j))
			}

			// Close the writer
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage,
				[]byte{}); err != nil {
				return
			}
		}
	}
}
