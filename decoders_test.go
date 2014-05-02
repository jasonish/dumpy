package main

import (
	"fmt"
	"testing"
	"time"
)

func TestSnortFastEventDecoder(t *testing.T) {

	eventBuffer := "11/15-22:56:29.943914  [**] [1:498:8] INDICATOR-COMPROMISE id check returned root [**] [Classification: Potentially Bad Traffic] [Priority: 2] {TCP} 217.160.51.31:80 -> 172.16.1.11:33189"

	event, err := DecodeEvent(eventBuffer)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, event.Timestamp,
		fmt.Sprintf("%d-11-15T22:56:29.943914", time.Now().Year()))
	AssertEquals(t, event.Protocol, "TCP")
	AssertEquals(t, event.SourceAddr, "217.160.51.31")
	AssertEquals(t, event.SourcePort, (uint16)(80))
	AssertEquals(t, event.DestAddr, "172.16.1.11")
	AssertEquals(t, event.DestPort, (uint16)(33189))
}

// Test a Suricata "fast" event which includes the year.
func TestSnortFastEventDecoderSuricata(t *testing.T) {

	eventBuffer := "11/15/2012-22:56:29.943914  [**] [1:498:8] INDICATOR-COMPROMISE id check returned root [**] [Classification: Potentially Bad Traffic] [Priority: 2] {TCP} 217.160.51.31:80 -> 172.16.1.11:33189"

	event, err := DecodeEvent(eventBuffer)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, event.Timestamp, "2012-11-15T22:56:29.943914")
	AssertEquals(t, event.Protocol, "TCP")
	AssertEquals(t, event.SourceAddr, "217.160.51.31")
	AssertEquals(t, event.SourcePort, (uint16)(80))
	AssertEquals(t, event.DestAddr, "172.16.1.11")
	AssertEquals(t, event.DestPort, (uint16)(33189))
}

func TestSuricataJsonEventDecoder(t *testing.T) {

	eventBuffer := `{"timestamp":"2014-05-05T13:34:35.453100","event_type":"alert","src_ip":"10.16.1.193","src_port":17500,"dest_ip":"255.255.255.255","dest_port":18500,"proto":"UDP","alert":{"action":"allowed","gid":1,"signature_id":2012648,"rev":3,"signature":"ET POLICY Dropbox Client Broadcasting","category":"Potential Corporate Privacy Violation","severity":1}}`

	event, err := DecodeEvent(eventBuffer)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, event.Timestamp, "2014-05-05T13:34:35.453100")
	AssertEquals(t, event.Protocol, "UDP")
	AssertEquals(t, event.SourceAddr, "10.16.1.193")
	AssertEquals(t, event.SourcePort, (uint16)(17500))
	AssertEquals(t, event.DestAddr, "255.255.255.255")
	AssertEquals(t, event.DestPort, (uint16)(18500))

}
