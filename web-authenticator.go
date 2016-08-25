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
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"github.com/gorilla/context"

	"golang.org/x/crypto/bcrypt"
)

type HandlerWrapper struct {
	authenticator *Authenticator
	handler       http.Handler
}

func (hw *HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hw.authenticator.authenticate(r) {
		hw.handler.ServeHTTP(w, r)
	} else {
		w.Header().Add("WWW-Authenticate", "Basic realm=restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
	}
}

type Authenticator struct {
	users map[string]string
}

func NewAuthenticator(config *Config) *Authenticator {
	if len(config.Users) == 0 {
		logger.Printf("WARNING: No users configuration. Authentication disabled.")
	}
	return &Authenticator{config.Users}
}

func (a *Authenticator) GetUsernameAndPassword(authHeader string) (string, string) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 {
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err == nil {
			userinfo := strings.SplitN(string(decoded), ":", 2)
			if len(userinfo) == 2 {
				return userinfo[0], userinfo[1]
			}
		}
	}

	return "", ""
}

func (a *Authenticator) CheckUsernameAndPassword(username string, password string) bool {
	hashedPassword, ok := a.users[username]
	if !ok {
		log.Printf("authentication error: user %s does not exist", username)
	} else {
		err := bcrypt.CompareHashAndPassword(
			([]byte)(hashedPassword), ([]byte)(password))
		if err == nil {
			return true
		}
		log.Printf("authentication error: bad password for user %s", username)
	}
	return false
}

func (a *Authenticator) authenticate(request *http.Request) bool {
	if len(a.users) == 0 {
		context.Set(request, "username", "<anonymous>")
		return true
	}
	authHeader := request.Header.Get("authorization")
	if authHeader != "" {
		username, password := a.GetUsernameAndPassword(authHeader)
		if username != "" {
			if a.CheckUsernameAndPassword(username, password) {
				context.Set(request, "username", username)
				return true
			}
		}
	}
	return false
}

func (a *Authenticator) WrapHandlerFunc(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		if a.authenticate(r) {
			handler(w, r)
		} else {
			w.Header().Add("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
		}
	}
	return wrapper
}

func (a *Authenticator) WrapHandler(handler http.Handler) http.Handler {
	return &HandlerWrapper{a, handler}
}
