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
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"
)

var (
	parsers_debug = true

	snortFastTimestampPattern = regexp.MustCompile("^(?P<month>\\d\\d)\\/(?P<day>\\d\\d)(?:\\/)?(?P<year>\\d{4})?-(?P<hour>\\d\\d):(?P<minute>\\d\\d):(?P<seconds>\\d\\d).(?P<microseconds>\\d+)")
	snortFastEventRegexp = regexp.MustCompile("{(?P<protocol>\\d+|\\w+)}\\s([\\d\\.]+):?(\\d+)?\\s..\\s([\\d\\.]+):?(\\d+)?")
)

type SuricataEveFlow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type SuricataJsonEvent struct {
	Timestamp  string `json:"timestamp"`
	Protocol   string `json:"proto"`
	SourceAddr string `json:"src_ip"`
	SourcePort uint16 `json:"src_port"`
	DestAddr   string `json:"dest_ip"`
	DestPort   uint16 `json:"dest_port"`
	EventType  string `json:"event_type"`

	Flow       SuricataEveFlow `json:"flow"`
}

type Event struct {
	Timestamp  string
	Protocol   string
	SourceAddr string
	SourcePort uint16
	DestAddr   string
	DestPort   uint16

	EventType  string
	Flow       SuricataEveFlow

	// The original event.
	Original   string
}

func (e *Event) ToPcapFilter() string {

	if e.SourcePort > 0 && e.DestPort > 0 {
		return fmt.Sprintf(
			"proto %s and ((host %s and port %d) and (host %s and port %d))",
			e.Protocol,
			e.SourceAddr, e.SourcePort,
			e.DestAddr, e.DestPort)
	} else {
		return fmt.Sprintf(
			"proto %s and host %s and host %s",
			e.Protocol, e.SourceAddr, e.DestAddr)
	}
}

func DecodeSnortFastEventTimestamp(buf string) string {

	matches := snortFastTimestampPattern.FindStringSubmatch(buf)
	if matches == nil {
		return ""
	}
	year := matches[3]
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	return fmt.Sprintf("%s-%s-%sT%s:%s:%s.%s",
		year, matches[1], matches[2],
		matches[4], matches[5], matches[6], matches[7])
}

// DecodeSnortFastEvent will decode Snort and Suricata "fast" style
// events.
func DecodeSnortFastEvent(buf string) *Event {

	eventMatches := snortFastEventRegexp.FindStringSubmatch(buf)
	if eventMatches == nil {
		if parsers_debug {
			log.Printf("event did not match snortFastEventRegexp")
		}
		return nil
	}
	timestamp := DecodeSnortFastEventTimestamp(buf)
	if timestamp == "" {
		if parsers_debug {
			log.Printf("failed to decode snort fast timestamp")
		}
		return nil
	}
	sourcePort, err := strconv.ParseUint(eventMatches[3], 10, 16)
	if err != nil {
		if parsers_debug {
			log.Print(err)
		}
		return nil
	}
	destPort, err := strconv.ParseUint(eventMatches[5], 10, 16)
	if err != nil {
		if parsers_debug {
			log.Print(err)
		}
		return nil
	}
	return &Event{
		Timestamp:  timestamp,
		Protocol:   eventMatches[1],
		SourceAddr: eventMatches[2],
		SourcePort: (uint16)(sourcePort),
		DestAddr:   eventMatches[4],
		DestPort:   (uint16)(destPort),
		Original:   buf,
	}
}

// DecodeSuricataJsonEvent decodes a Suricata style JSON event.
func DecodeSuricataJsonEvent(buf string) *Event {

	suricataJsonEvent := SuricataJsonEvent{}

	log.Println("** Decoding...")

	if err := json.Unmarshal(([]byte)(buf), &suricataJsonEvent); err != nil {
		if parsers_debug {
			log.Print(err)
		}
		return nil
	}
	log.Println(suricataJsonEvent.EventType)
	log.Println(suricataJsonEvent.Flow)

	return &Event{
		Timestamp:  suricataJsonEvent.Timestamp,
		Protocol:   suricataJsonEvent.Protocol,
		SourceAddr: suricataJsonEvent.SourceAddr,
		SourcePort: suricataJsonEvent.SourcePort,
		DestAddr:   suricataJsonEvent.DestAddr,
		DestPort:   suricataJsonEvent.DestPort,
		EventType: suricataJsonEvent.EventType,

		Flow: suricataJsonEvent.Flow,

		Original:   buf,
	}
}

func DecodeEvent(buf string) (*Event, error) {
	if event := DecodeSuricataJsonEvent(buf); event != nil {
		return event, nil
	}
	if event := DecodeSnortFastEvent(buf); event != nil {
		return event, nil
	}
	return nil, fmt.Errorf("error: event not recognized by any decoders")
}
