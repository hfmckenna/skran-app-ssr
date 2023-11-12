package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		// Apple listens on 5000, AWS needs 5000
		if runtime.GOOS == "darwin" {
			port = "8080"
		} else {
			port = "5000"
		}
	}

	f, _ := os.Create("/var/log/golang/golang-server.log")
	defer f.Close()
	log.SetOutput(f)

	const indexPage = "templates/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(indexPage))
		tmpl.Execute(w, nil)
	})

	log.Printf("Listening on port %s\n\n", port)
	http.ListenAndServe(":"+port, nil)
}
