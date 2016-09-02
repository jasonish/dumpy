// Copyright (c) 2014 Jason Ish. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above
//    copyright notice, this list of conditions and the following
//    disclaimer in the documentation and/or other materials provided
//    with the distribution.
//
// THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED
// WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"encoding/json"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/jasonish/dumpy/config"
	"golang.org/x/net/context"
)

func HttpErrorAndLog(w http.ResponseWriter, r *http.Request, code int, format string, v ...interface{}) {
	error := fmt.Sprintf(format, v...)
	logger.PrintWithRequest(r, error)
	http.Error(w, error, code)
}

type HttpError struct {
	Message string
	Code    int
}

func (he *HttpError) Error() string {
	return he.Message
}

func JsonResponse(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

func VersionHandlerFunc(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{} {
		"version": VERSION,
	}
	JsonResponse(w, version)
}

type IndexHandler struct {
	config *config.Config
}

func (h IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	preparedEvent := r.FormValue("event")

	box, err := rice.FindBox("www")
	if err != nil {
		log.Fatal(err)
	}
	indexString, err := box.String("index.html")
	if err != nil {
		log.Fatal(err)
	}

	model := map[string]interface{}{
		"spools": h.config.Spools,
		"event":  preparedEvent,
	}

	templatePage, err := template.New("index").Parse(indexString)
	templatePage.Execute(w, model)
}

// Authentication middleware implemented as a handler function.
func AuthMiddlewareHandlerFunc(authenticator *Authenticator, handleFunc http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		user := authenticator.AuthenticateHttpRequest(r)
		if user != nil {
			handleFunc(w, r.WithContext(context.WithValue(r.Context(), "User", user)))
		} else {
			w.Header().Add("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
		}

	}

}

// Authentication middleware implemented as an http.Handler.
func AuthMiddlewareHandler(authenticator *Authenticator, h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := authenticator.AuthenticateHttpRequest(r)
		if user != nil {
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "User", user)))
		} else {
			w.Header().Add("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
		}

	})

}

func StartServer(config *config.Config) {

	authenticator := NewAuthenticator(config)
	router := mux.NewRouter()

	// Legacy fetch handler.
	router.Handle("/fetch", &FetchHandler{config})

	// Setup API v1 handlers.
	ApiV1SetupRoutes(config, router.PathPrefix("/api/1").Subrouter())

	router.HandleFunc("/version", VersionHandlerFunc)

	router.Handle("/", &IndexHandler{config})

	router.PathPrefix("/").Handler(
		http.FileServer(rice.MustFindBox("www").HTTPBox()))

	http.Handle("/", AuthMiddlewareHandler(authenticator, router))

	addr := fmt.Sprintf(":%d", config.Port)
	if !config.Tls.Enabled {
		log.Printf("Starting server on %s", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Starting server on %s with TLS", addr)
		err := http.ListenAndServeTLS(addr, config.Tls.Certificate,
			config.Tls.Key, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}
