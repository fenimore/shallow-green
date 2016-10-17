package main

import (
	"fmt"
	"net/http"
)

// Index page, link to new game
func Index(w http.ResponseWriter,
	r *http.Request) {
	html := `
<html>
<link href="/css/style.css" rel="stylesheet">
<h1>Ghess Index</h1>
<a href=/new/black >New Game Computer Vs Human</a><br>
<a href=/new/white >New Game Human Vs Computer</a><br>
<hr>
<a href=/board >View Current Game</a><br>
<br><br><br>
<a href="https://github.com/polypmer/go-chess">Source Code</a>
</html>
`
	fmt.Fprintln(w, html)
}

func NewGame(w http.ResponseWriter,
	r *http.Request) {
	// new Board
	// Make first move if black
}

func ViewGame(w http.ResponseWriter,
	r *http.Request) {

}

func PlayGame(w http.ResponseWriter,
	r *http.Request) {

}

/* AI Game */
// TODO:
