package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	f, _ := os.Create("/var/log/golang/golang-server.log")
	defer f.Close()
	log.SetOutput(f)

	const indexPage = "index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexPage)
	})

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
