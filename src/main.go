package main

import (
	"log"
	"net/http"
	"os"
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

	http.HandleFunc("/", home)

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
