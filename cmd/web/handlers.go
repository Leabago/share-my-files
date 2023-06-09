package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.tmpl.html", &templateData{})
}

func (app *application) homeGetFiles(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("homeGetFiles")

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	if mediaType == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				log.Fatal(err)
			}
			// fmt.Printf("Part %q: %q\n", p.Header.Get("Foo"), slurp)

			fmt.Printf("FileName %s: \n", p.FileName())

			path := "/tmp/savedFiles" + p.FileName()
			err = os.WriteFile(path, slurp, 0644)
			check(err)

		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
