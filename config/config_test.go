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
