package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"skran-app-ssr/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	assets := os.Getenv("ASSETS_DOMAIN")
	if assets == "" {
		http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	}

	homeHandler := wrap(api.Home)

	http.HandleFunc("/", homeHandler)

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type wrapper func(io.Writer)
type httpHandler func(http.ResponseWriter, *http.Request)

func wrap(function wrapper) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		function(w)
	}
}
