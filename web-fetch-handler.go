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
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// FetchHandler is the HTTP handler for "fetch" (download) requests of
// captured packets.
//
// The request can be a GET or POST of url encoded parameters.
//
// Parameters:
//
//   filter - PCAP filter or event string.
//
//   start-time - RFC3339 formatted date for packet capture to start.
//
//   duration - String specifying duration (eg: 1m, 3h).
//
//   duration-before - If event, the duration before the timestamp in
//       the event.
//
//   duration-after - If event, the duration after the timestamp in
//       the event.
//
//   spool - The name of the spool to look for packets.
//
//   timezone-offset - The timezone offset in the format of {+-}HH:MM
//       to be used if not timezone offset is present in the filter
//       string.
type FetchHandler struct {
	config *Config
}

func (h *FetchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	queryType := r.FormValue("query-type")

	switch queryType {
	case "pcap-filter":
		h.HandleFilterRequest(w, r)
	case "event":
		h.HandleEventRequest(w, r)
	case "":
		HttpErrorAndLog(w, r, http.StatusBadRequest,
			"Missing query-type parameter.")
	default:
		HttpErrorAndLog(w, r, http.StatusBadRequest,
			"Unknown query-type: %s", queryType)
	}
}

func (h *FetchHandler) HandleFilterRequest(w http.ResponseWriter, r *http.Request) {

	filter := r.FormValue("filter")
	argStartTime := r.FormValue("start-time")
	argDuration := r.FormValue("duration")
	argSpool := r.FormValue("spool")

	if argStartTime == "" || argDuration == "" {
		http.Error(w, "start time and duration required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, argStartTime)
	if err != nil {
		http.Error(w, "failed to parse time: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(argDuration)
	if err != nil {
		http.Error(w, "failed to parse duraction: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	if argSpool == "" {
		http.Error(w, "spool name required", http.StatusBadRequest)
		return
	}

	spool := h.config.GetSpoolByName(argSpool)
	if spool == nil {
		http.Error(w, fmt.Sprintf("spool %s not configured", argSpool),
			http.StatusInternalServerError)
		return
	}

	logger.PrintfWithRequest(r, "Preparing dumper request: %s",
		map[string]string{
			"start-time":  startTime.String(),
			"duration":    duration.String(),
			"pcap-filter": filter,
		})

	dumperArgs := []string{
		"-directory", spool.Directory,
		"-prefix", spool.Prefix,
		"-start-time", strconv.FormatInt(startTime.Unix(), 10),
		"-duration", strconv.FormatInt(int64(duration.Seconds()), 10),
		"-filter", filter,
	}
	h.RunDumper(w, r, dumperArgs, 0)
}

func (h *FetchHandler) HandleEventRequest(w http.ResponseWriter, r *http.Request) {

	event, err := DecodeEvent(r.FormValue("event"))
	if err != nil {
		HttpErrorAndLog(w, r, http.StatusBadRequest, err.Error())
		return
	}

	eventTimestamp := ParseTime(event.Timestamp,
		r.FormValue("default-timezone-offset"))
	if eventTimestamp == nil {
		http.Error(w, "failed to parse timestamp: "+event.Timestamp,
			http.StatusBadRequest)
		return
	}

	durationBefore, err := time.ParseDuration(
		"-" + r.FormValue("duration-before"))
	if err != nil {
		http.Error(w, fmt.Sprintf("duration-before: %s", err),
			http.StatusBadRequest)
		return
	}

	durationAfter, err := time.ParseDuration(r.FormValue("duration-after"))
	if err != nil {
		http.Error(w, fmt.Sprintf("duration-before: %s", err),
			http.StatusBadRequest)
		return
	}

	startTime := eventTimestamp.Add(durationBefore)
	endTime := eventTimestamp.Add(durationAfter)
	duration := endTime.Sub(startTime)

	spool := h.config.GetSpoolByName(r.FormValue("spool"))
	if spool == nil {
		http.Error(w,
			fmt.Sprintf("spool %s not configured", r.FormValue("spool")),
			http.StatusInternalServerError)
		return
	}

	logger.PrintfWithRequest(r, "Preparing dumper request: %s",
		map[string]string{
			"start-time":  startTime.String(),
			"duration":    duration.String(),
			"pcap-filter": event.ToPcapFilter(),
			"event":       event.Original,
		})

	dumperArgs := []string{
		"-directory", spool.Directory,
		"-prefix", spool.Prefix,
		"-start-time", strconv.FormatInt(startTime.Unix(), 10),
		"-duration", strconv.FormatInt(int64(duration.Seconds()), 10),
		"-filter", event.ToPcapFilter(),
	}
	h.RunDumper(w, r, dumperArgs, event.SignatureId)
}

func (h *FetchHandler) RunDumper(w http.ResponseWriter, r *http.Request, args []string, signature_id uint32) {
	dumper := exec.Command(os.Args[0], "dump")
	dumper.Args = append(dumper.Args, args...)

	stdout, err := dumper.StdoutPipe()
	if err != nil {
		http.Error(w, "failed to attach pipe: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	stderr, err := dumper.StderrPipe()
	if err != nil {
		http.Error(w, "failed to attach stderr pipe: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// For logging.
	go func() {
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			logger.Printf("dumpy dumper [%d] (stderr) %s", dumper.Process.Pid, line)
		}
	}()

	err = dumper.Start()
	if err != nil {
		HttpErrorAndLog(w, r, http.StatusInternalServerError,
			"failed to execute dumper: %s", err)
		return
	}
	logger.PrintfWithRequest(r, "dumper with pid %d started: %s",
		dumper.Process.Pid, dumper.Args)

	bytesWritten := 0
	for {
		buf := make([]byte, 8192)
		n, err := stdout.Read(buf)
		if err != nil {
			if err != io.EOF {
				logger.PrintfWithRequest(r, "unexpected dump error: %s", err)
			}
			break
		}

		if bytesWritten == 0 {
			var cdstr string
			if signature_id != 0 {
				cdstr = fmt.Sprintf("attachment; filename=%d.pcap", signature_id)
			} else {
				cdstr = "attachment; filename=dumpy.pcap"
			}
			w.Header().Add("content-type", "application/vnd.tcpdump.pcap")
			w.Header().Add("content-disposition", cdstr)
		}
		n, err = w.Write(buf[0:n])
		if err != nil {
			logger.PrintfWithRequest(r,
				"Write failed; client may have disconnected: %s", err)
			break
		}
		bytesWritten += n
	}

	if bytesWritten == 0 {
		http.Error(w, "No packets found.", http.StatusNotFound)
	} else {
		logger.PrintfWithRequest(r, "Wrote %d bytes of packet data",
			bytesWritten)
	}
}
