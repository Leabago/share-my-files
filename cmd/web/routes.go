package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, app.secureHeaders)
	dynamicMiddleware := alice.New()

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.redirectHome))
	mux.Get("/upload", dynamicMiddleware.ThenFunc(app.createSnippetForm))
	mux.Post("/upload", dynamicMiddleware.ThenFunc(app.homeGetFiles))
	mux.Post("/archive", dynamicMiddleware.ThenFunc(app.archive))
	mux.Get("/archive/:id", dynamicMiddleware.ThenFunc(app.showSnippet))
	mux.Get("/archive/download/:id", dynamicMiddleware.ThenFunc(app.getSnippet))

	// delte one file from list
	mux.Post("/delete/:name", dynamicMiddleware.ThenFunc(app.deleteOneFile))
	mux.Get("/download", dynamicMiddleware.ThenFunc(app.createDownloadForm))

	// ping
	mux.Get("/ping", http.HandlerFunc(ping))

	// static
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static/", fileServer))

	return standardMiddleware.Then(mux)
}
