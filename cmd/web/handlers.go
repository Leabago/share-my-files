package main

import (
	"fmt"
	"net/http"
	"share-my-file/pkg/forms"
	"share-my-file/pkg/models"
	"strconv"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl.html", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get(":id")

	var file = &models.File{
		FolderCode: code,
	}

	// s, err := app.files.Get()

	// if err != nil {
	// 	if errors.Is(err, models.ErrNoRecord) {
	// 		app.notFound(w)
	// 	} else {
	// 		app.serverError(w, err)
	// 	}
	// 	return
	// }

	app.render(w, r, "show.page.tmpl.html", &templateData{
		File: file,
	})

	// filePath := folderPath + folderBegin + code + zipName
	// filename := folderBegin + code + zipName

	// w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	// w.Header().Set("Content-Type", "application/octet-stream")
	// http.ServeFile(w, r, filePath)

}

func (app *application) getSnippet(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get(":id")

	// var file = &models.File{
	// 	FolderCode: code,
	// }

	// app.render(w, r, "show.page.tmpl.html", &templateData{
	// 	File: file,
	// })

	filePath := folderPath + folderBegin + code + zipName
	filename := folderBegin + code + zipName

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)

}

func (app *application) homeGetFiles(w http.ResponseWriter, r *http.Request) {

	// get folder name
	var code = createUserCode()
	var zipFileName = app.createFoldeWithCoderForFiles(code)
	app.logger.infoLog.Printf("create new folder %s", zipFileName)
	ParseMediaType(r, zipFileName)
	http.Redirect(w, r, fmt.Sprintf("/snippet/%s", code), http.StatusSeeOther)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
