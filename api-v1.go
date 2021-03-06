// Copyright (c) 2016 Jason Ish. All rights reserved.
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
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	"github.com/jasonish/dumpy/dumper"
	"time"
	"github.com/jasonish/dumpy/config"
	"errors"
	"fmt"
	"github.com/jasonish/dumpy/env"
)

func ParseTimestamp(timestamp string) (int64, error) {
	asInt64, err := strconv.ParseInt(timestamp, 10, 64)
	if err == nil {
		return asInt64, nil
	}
	asTime, err := ParseTime(timestamp)
	if err == nil {
		return asTime.Unix(), nil
	}
	return 0, errors.New(fmt.Sprintf("Failed to parse timestamp: %s", timestamp))
}

func ApiV1DownloadRequestHandler(env env.Env, w http.ResponseWriter, r *http.Request) error {
	dumperOptions := dumper.DumperOptions{}
	var err error
	filename := "dumpy.pcap"

	if r.FormValue("duration") != "" && r.FormValue("endTime") != "" {
		return &HttpError{"Duration and endTime cannot both be provided", http.StatusBadRequest}
	}

	if r.FormValue("startTime") != "" {
		dumperOptions.StartTime, err = ParseTimestamp(r.FormValue("startTime"))
		if err != nil {
			return &HttpError{err.Error(), http.StatusBadRequest}
		}
	}

	if r.FormValue("endTime") != "" {
		endTime, err := ParseTimestamp(r.FormValue("endTime"))
		if err != nil {
			return &HttpError{err.Error(), http.StatusBadRequest}
		}
		duration := endTime - dumperOptions.StartTime
		if duration < 0 {
			return &HttpError{"endTime cannot be before startTime", http.StatusBadRequest}
		}
		dumperOptions.Duration = duration
	}

	if r.FormValue("duration") != "" {
		duration, err := time.ParseDuration(r.FormValue("duration"))
		if err != nil {
			return &HttpError{"Failed to parse duration", http.StatusBadRequest}
		}
		dumperOptions.Duration = int64(duration.Seconds())
	}

	if r.FormValue("filter") != "" {
		dumperOptions.Filter = r.FormValue("filter")
	}

	// Filename to be used in the content-disposition.
	if r.FormValue("filename") != "" {
		filename = r.FormValue("filename")
	}

	var spool *config.SpoolConfig

	if r.FormValue("spool") != "" {
		spool = env.Config.GetSpoolByName(r.FormValue("spool"))
		if spool == nil {
			return &HttpError{"Spool not found: " + r.FormValue("spool"), http.StatusBadRequest}
		}
	} else {
		spool = env.Config.GetFirstSpool()
	}

	if spool == nil {
		return &HttpError{"No spools found.", http.StatusBadRequest}
	} else {
		dumperOptions.Directory = spool.Directory
		dumperOptions.Prefix = spool.Prefix
		dumperOptions.Recursive = spool.Recursive
	}

	proxy := env.ProxyCreator.NewProxy(spool, dumperOptions, w, filename)
	proxy.Run()

	return nil
}

type ApiV1Handler struct {
	env     env.Env
	handler func(env env.Env, w http.ResponseWriter, r *http.Request) error
}

func (h ApiV1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	err := h.handler(h.env, w, r)
	if err != nil {
		switch t := err.(type) {
		case *HttpError:
			http.Error(w, t.Message, t.Code)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

}

func ApiV1SetupRoutes(env env.Env, router *mux.Router) {
	router.Handle("/download", ApiV1Handler{env, ApiV1DownloadRequestHandler})
}
