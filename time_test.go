// Copyright (c) 2014-2016 Jason Ish. All rights reserved.
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
	"time"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestParseTime(t *testing.T) {

	// Basic RFC3339 value at UTC.
	value, err := ParseTime("2014-04-30T21:54:09Z")
	FailIfNotNil(t, err)
	FailIfNotEqual(t, int64(1398894849), value.Unix())

	// With timezone offsets.
	value, err = ParseTime("2014-04-30T15:56:42-06:00")
	if err != nil {
		t.Fatal(err)
	}
	if value.Unix() != 1398895002 {
		t.Fatalf("expected 1398895002, got %d", value.Unix())
	}

	// With modified timezone offset.
	value, err = ParseTime("2014-04-30T15:56:42.857989-0600")
	if err != nil {
		t.Fatal(err)
	}
	if value.Unix() != 1398895002 {
		t.Fatalf("expected 1398895002, got %d", value.Unix())
	}

	// With microsends, UTC.
	value, err = ParseTime("2014-04-30T15:56:42.857989Z")
	if err != nil {
		t.Fatal(err)
	}
	if value.Unix() != 1398873402 {
		t.Fatalf("expected 1398873402, got %d", value.Unix())
	}
}

// Test durations the UI may present, or we provide in the
// documentation.
func TestParseDuration(t *testing.T) {

	var value time.Duration
	var err error

	value, err = time.ParseDuration("1m")
	FailIfNotNil(t, err)
	FailIfNotEqual(t, int64(value.Seconds()), int64(60))

	value, err = time.ParseDuration("15s")
	FailIfNotNil(t, err)
	FailIfNotEqual(t, int64(value.Seconds()), int64(15))

	value, err = time.ParseDuration("1h")
	FailIfNotNil(t, err)
	FailIfNotEqual(t, int64(value.Seconds()), int64(3600))
}
