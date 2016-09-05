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

// APIv1 test code. This is a bit of a mess right now.

package main

import (
	"github.com/jasonish/dumpy/config"
	"github.com/jasonish/dumpy/dumper"
	"github.com/jasonish/dumpy/env"
	"github.com/jasonish/dumpy/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type MockProxy struct {
	options  dumper.DumperOptions
	filename string
	didRun   bool
}

func (p *MockProxy) Run() {
	p.didRun = true
}

type MockProxyDumpCreator struct {
	proxy *MockProxy
}

func (c *MockProxyDumpCreator) NewProxy(spoolConfig *config.SpoolConfig,
	options dumper.DumperOptions, w http.ResponseWriter, filename string) dumper.Proxy {
	c.proxy = &MockProxy{options: options,
		filename: filename}
	return c.proxy
}

func (c *MockProxyDumpCreator) GetProxy() *MockProxy {
	return c.proxy
}

func apiV1RequestWrapper(env env.Env, w http.ResponseWriter, r *http.Request,
	fn func(env env.Env, w http.ResponseWriter, r *http.Request) error) *HttpError {
	err := fn(env, w, r)
	if err != nil {
		return err.(*HttpError)
	}
	return nil
}

func TestApiV1TestDownloadValid(t *testing.T) {
	env := env.New()
	mockProxyCreator := MockProxyDumpCreator{}
	env.ProxyCreator = &mockProxyCreator

	// Create an emtpy spool to be used as the default.
	env.Config.Spools = []*config.SpoolConfig{
		&config.SpoolConfig{},
	}

	// endTime and duration not allowed together.
	request := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/api/v1/download"},
		Form: url.Values{
			"startTime": {"0"},
			"duration":  {"1s"},
		},
	}

	rr := httptest.NewRecorder()
	err := apiV1RequestWrapper(env, rr, request, ApiV1DownloadRequestHandler)
	test.FailIf(t, err != nil)
	proxy := mockProxyCreator.GetProxy()
	test.FailIfNot(t, proxy.didRun)
	test.FailIf(t, proxy.options.StartTime != 0)
	test.FailIf(t, proxy.options.Duration != 1)
}

// Test that if endTime and duration are provided that we fail the request.
func TestApiV1DownloadRequestHandler(t *testing.T) {
	env := env.New()
	env.ProxyCreator = &MockProxyDumpCreator{}

	// endTime and duration not allowed together.
	request := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/api/v1/download"},
		Form: url.Values{
			"endTime":  {"1"},
			"duration": {"1s"},
		},
	}

	rr := httptest.NewRecorder()
	err := apiV1RequestWrapper(env, rr, request, ApiV1DownloadRequestHandler)
	if err == nil {
		t.Fatal("err should not be nil")
	}
	if err.Code != http.StatusBadRequest {
		t.Fatal("unexpected error code")
	}

}
