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

package config

import (
	"encoding/json"
	"testing"
)

func AssertEquals(t *testing.T, value interface{}, expected interface{}) {
	if value != expected {
		if _, ok := value.(int64); ok {
			t.Fatalf("got %d, expected %d", value, expected)
		} else {
			// Fall back to strings.
			t.Fatalf("got %s, expected %s", value, expected)
		}
	}
}

func TestConfigGetSpoolByName(t *testing.T) {
	jsonConfig := `{
  "spools": [
    {"name": "first",
     "directory": "/capture/first",
     "prefix": "first.pcap"},
    {"name": "second",
     "directory": "/capture/second",
     "prefix": "second.pcap"}
  ]
}`

	_config := Config{}
	err := json.Unmarshal(([]byte)(jsonConfig), &_config)
	if err != nil {
		t.Fatal(err)
	}

	spool := _config.GetSpoolByName("first")
	if spool == nil {
		t.Fatal("spool first not found")
	}

	AssertEquals(t, spool.Name, "first")
	AssertEquals(t, spool.Directory, "/capture/first")
	AssertEquals(t, spool.Prefix, "first.pcap")

	spool = _config.GetSpoolByName("second")
	if spool == nil {
		t.Fatal("spool second not found")
	}
	AssertEquals(t, spool.Name, "second")
	AssertEquals(t, spool.Directory, "/capture/second")
	AssertEquals(t, spool.Prefix, "second.pcap")

	spool = _config.GetSpoolByName("third")
	if err != nil {
		t.Fatalf("expected nil instead of a spool")
	}

}
