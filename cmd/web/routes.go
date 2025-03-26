package main

import (
	"net/http"
	"strconv"
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

func (app *application) routes() (http.Handler, error) {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, app.secureHeaders)
	mux := pat.New()
	var dynamicMiddleware alice.Chain

	dnsValidation := getEnv("DNS_VALIDATION", app.logger)

	dnsValidBool, err := strconv.ParseBool(dnsValidation)
	if err != nil {
		return standardMiddleware.Then(mux), err
	}

	if dnsValidBool {
		dynamicMiddleware = alice.New(app.dnsValidation)
	} else {
		dynamicMiddleware = alice.New()
	}

	mux.Get("/kek", dynamicMiddleware.ThenFunc(app.healthzHandler))

	mux.Get("/", dynamicMiddleware.ThenFunc(app.redirectHome))
	mux.Get("/upload", dynamicMiddleware.ThenFunc(app.createSnippetForm))
	mux.Post("/upload", dynamicMiddleware.ThenFunc(app.homeGetFiles))
	mux.Post("/archive", dynamicMiddleware.ThenFunc(app.archive))
	mux.Get("/archive/:id", dynamicMiddleware.ThenFunc(app.showSnippet))
	mux.Get("/archive/download/:id", dynamicMiddleware.ThenFunc(app.getSnippet))

	mux.Get("/user-code", dynamicMiddleware.ThenFunc(app.getUserCode))

	// delte one file from list
	mux.Del("/delete/:name", dynamicMiddleware.ThenFunc(app.deleteOneFile))

	// find file by sessionCode
	mux.Get("/download", dynamicMiddleware.ThenFunc(app.createDownloadForm))

	// HTTP Liveness and Readiness Probes
	mux.Get("/healthz", dynamicMiddleware.ThenFunc(app.healthzHandler))
	mux.Get("/readyz", dynamicMiddleware.ThenFunc(app.readyzHandler))

	// HTTP Liveness and Readiness Probes
	mux.Get("/health", dynamicMiddleware.ThenFunc(app.healthzHandler))

	// ping
	mux.Get("/ping", http.HandlerFunc(ping))

	// static
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static/", fileServer))

	// https, validation for ZeroSSL
	fileServerSSL := http.FileServer(http.Dir("./ssl"))
	mux.Get("/.well-known/pki-validation/", http.StripPrefix("/.well-known/pki-validation/", fileServerSSL))

	return standardMiddleware.Then(mux), nil
}
