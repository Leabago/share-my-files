package main

import (
	"net/http"
	"sync"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// SessionState tracks the state of each session
type SessionState struct {
	IsArchiving bool
}

var (
	sessionMap = make(map[string]*SessionState) // Maps session_id to SessionState
	mutex      = &sync.Mutex{}                  // Protects access to sessionMap
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

	mux.Get("/user-code", dynamicMiddleware.ThenFunc(app.getUserCode))

	// delte one file from list
	mux.Del("/delete/:name", dynamicMiddleware.ThenFunc(app.deleteOneFile))
	mux.Get("/download", dynamicMiddleware.ThenFunc(app.createDownloadForm))

	// ping
	mux.Get("/ping", http.HandlerFunc(ping))

	// static
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static/", fileServer))

	return standardMiddleware.Then(mux)
}
