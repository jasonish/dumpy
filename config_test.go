package main

import (
	"encoding/json"
	"testing"
)

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

	config := Config{}
	err := json.Unmarshal(([]byte)(jsonConfig), &config)
	if err != nil {
		t.Fatal(err)
	}

	spool := config.GetSpoolByName("first")
	if spool == nil {
		t.Fatal("spool first not found")
	}
	AssertEquals(t, spool.Name, "first")
	AssertEquals(t, spool.Directory, "/capture/first")
	AssertEquals(t, spool.Prefix, "first.pcap")

	spool = config.GetSpoolByName("second")
	if spool == nil {
		t.Fatal("spool second not found")
	}
	AssertEquals(t, spool.Name, "second")
	AssertEquals(t, spool.Directory, "/capture/second")
	AssertEquals(t, spool.Prefix, "second.pcap")

	spool = config.GetSpoolByName("third")
	if err != nil {
		t.Fatalf("expected nil instead of a spool")
	}

}
