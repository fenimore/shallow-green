package main

import (
	"time"

	"github.com/polypmer/ghess"
)

type Game struct {
	g    ghess.Board
	date time.Time
}
