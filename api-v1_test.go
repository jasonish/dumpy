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
	"testing"
	"net/http"
	"github.com/jasonish/dumpy/config"
	"net/http/httptest"
	"net/url"
)

func TestApiV1DownloadRequestHandler(t *testing.T) {

	// Wrapper function to convert the error returned from a handler to a *HttpError.
	errorConverter := func(c *config.Config, w http.ResponseWriter, r *http.Request, fn func(c *config.Config, w http.ResponseWriter, r *http.Request) error) *HttpError {
		err := fn(c, w, r)
		httpError := err.(*HttpError)
		return httpError
	}

	doSomething(&DumperProxy{})

	c := config.NewConfig()

	// endTime and duration not allowed together.
	request := &http.Request{
		Method: "GET",
		URL: &url.URL{Path: "/api/v1/download"},
		Form: url.Values{
			"endTime": {"1"},
			"duration": {"1s"},
		},
	}

	rr := httptest.NewRecorder()
	err := errorConverter(c, rr, request, ApiV1DownloadRequestHandler)
	if err == nil {
		t.Fatal("err should not be nil")
	}
	if err.Code != http.StatusBadRequest {
		t.Fatal("unexpected error code")
	}
}