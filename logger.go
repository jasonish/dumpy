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
	"log"
	"net/http"
	"os"
	"strings"
)

// Logger is a wrapper around the standard logger that implements some
// http.Request aware functions.
type Logger struct {
	*log.Logger
}

func NewLogger(prefix string) *Logger {
	return &Logger{log.New(os.Stderr, prefix, log.Ldate | log.Ltime)}
}

func (l *Logger) PrintfWithRequest(r *http.Request, format string, v ...interface{}) {
	user := r.Context().Value("User").(*User)
	l.Printf("[%s@%s] %s", user.Username, l.getRemoteAddr(r),
		fmt.Sprintf(format, v...))
}

func (l *Logger) PrintWithRequest(r *http.Request, v ...interface{}) {
	user := r.Context().Value("User").(*User)
	l.Printf("[%s@%s] %s", user.Username, l.getRemoteAddr(r),
		fmt.Sprint(v...))
}

// getRemoteAddr gets the remote address of the request without the
// port information.
func (l *Logger) getRemoteAddr(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}
