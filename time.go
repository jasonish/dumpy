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
	"time"
	"errors"
)

// Time related functions.

const (
	// Just like RFC3339Nano but allow for a timezone offset without
	// the ":" between the hours and minutes.
	RFC3339Nano_Modified = "2006-01-02T15:04:05.999999999Z0700"
)

// Utility time parsing function that can handle the following formats:
//
// - 2014-04-30T21:54:09Z
// - 2014-04-30T15:56:42-06:00
// - 2014-04-30T15:56:42-0600
// - 2014-04-30T15:56:42.857989-0600
func ParseTime(value string) (time.Time, error) {

	// First try as RFC3339Nano.
	result, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return result, nil
	}

	// Then try out modified RFC3339Nano.
	result, err = time.Parse(RFC3339Nano_Modified, value)
	if err == nil {
		return result, nil
	}

	return time.Time{}, errors.New("Failed to parse time")
}
