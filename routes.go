package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	router.Handle("/", http.FileServer(http.Dir("./static/")))
	return router
}

// Define handlers in handlers.go
var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"NewAi",
		"GET",
		"/new/{player}",
		NewGame,
	},
	Route{
		"ViewAi",
		"GET",
		"/view/{id}",
		ViewGame,
	},
	Route{
		"PlayAi",
		"GET",
		"/play/{id}/{orig}/{dest}",
		PlayGame,
	},
	// New websockets
	// New ai game (ajax)
	// Show websockets
	// Show ai
	// response websockets
	// response ai
	// about?
}

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
