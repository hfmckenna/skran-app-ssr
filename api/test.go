package api

import (
	"html/template"
	"io"
	"log"
)

func Test(w io.Writer) {
	tmpl, _ := template.New("").Parse("Test thing!")
	err := tmpl.Execute(w, "")
	if err != nil {
		log.Fatal(err)
	}
}
