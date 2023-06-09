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

	// snippets
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Post("/", dynamicMiddleware.ThenFunc(app.homeGetFiles))

	// ping
	mux.Get("/ping", http.HandlerFunc(ping))

	// static
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static/", fileServer))

	return standardMiddleware.Then(mux)
}
