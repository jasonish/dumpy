package main

import (
	"fmt"
	"testing"
	"time"
)

func FailIf(t *testing.T, expression bool) {
	if expression {
		t.FailNow()
	}
}

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

func TestDececodSuricataEveFlowEvent(t *testing.T) {
	raw := `{
    "timestamp": "2016-08-30T21:25:54.000103-0600",
    "flow_id": 956293338,
    "event_type": "flow",
    "src_ip": "10.16.1.10",
    "src_port": 45744,
    "dest_ip": "216.17.8.3",
    "dest_port": 443,
    "proto": "TCP",
    "flow": {
      "pkts_toserver": 4,
      "pkts_toclient": 4,
      "bytes_toserver": 260,
      "bytes_toclient": 248,
      "start": "2016-08-30T21:24:31.075881-0600",
      "end": "2016-08-30T21:24:52.926643-0600",
      "age": 21,
      "state": "new",
      "reason": "timeout"
    },
    "tcp": {
      "tcp_flags": "00",
      "tcp_flags_ts": "00",
      "tcp_flags_tc": "00"
    },
    "host": "fw",
    "@version": "1",
    "@timestamp": "2016-08-31T03:25:54.000Z",
    "input_type": "log",
    "count": 1,
    "offset": 134731057,
    "source": "/var/log/suricata/eve.json",
    "type": "log",
    "fields": {
      "type": "eve"
    },
    "beat": {
      "hostname": "fw.unx.ca",
      "name": "fw.unx.ca"
    },
    "tags": [
      "beats_input_codec_json_applied"
    ],
    "geoip": {
      "ip": "216.17.8.3",
      "country_code2": "US",
      "country_code3": "USA",
      "country_name": "United States",
      "continent_code": "NA",
      "region_name": "MN",
      "city_name": "Plymouth",
      "postal_code": "55441",
      "latitude": 45.0059,
      "longitude": -93.4305,
      "dma_code": 613,
      "area_code": 763,
      "timezone": "America/Chicago",
      "real_region_name": "Minnesota",
      "location": [
        -93.4305,
        45.0059
      ],
      "coordinates": [
        -93.4305,
        45.0059
      ]
    }
  }`

	event, err := DecodeEvent(raw)
	FailIf(t, err != nil)
	FailIf(t, event == nil)
	FailIf(t, event.EventType != "flow")
}