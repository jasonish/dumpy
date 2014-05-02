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

	spool, err := config.GetSpoolByName("first")
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, spool.Name, "first")
	AssertEquals(t, spool.Directory, "/capture/first")
	AssertEquals(t, spool.Prefix, "first.pcap")

	spool, err = config.GetSpoolByName("second")
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, spool.Name, "second")
	AssertEquals(t, spool.Directory, "/capture/second")
	AssertEquals(t, spool.Prefix, "second.pcap")

	_, err = config.GetSpoolByName("third")
	if err == nil {
		t.Fatalf("expected error while getting non-existant spool by name")
	}

}
