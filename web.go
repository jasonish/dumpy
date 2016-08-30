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

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/jasonish/dumpy/config"
)

func HttpErrorAndLog(w http.ResponseWriter, r *http.Request, code int,
	format string, v ...interface{}) {

	error := fmt.Sprintf(format, v...)
	logger.PrintWithRequest(r, error)
	http.Error(w, error, code)
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

func StartServer(config *config.Config) {

	authenticator := NewAuthenticator(config)

	router := mux.NewRouter()

	router.Handle("/fetch", authenticator.WrapHandler(&FetchHandler{config}))
	router.Handle("/", authenticator.WrapHandler(&IndexHandler{config}))

	router.PathPrefix("/").Handler(
		http.FileServer(rice.MustFindBox("www").HTTPBox()))

	http.Handle("/", router)

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
