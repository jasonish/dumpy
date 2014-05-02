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
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"dumpy/libpcap"
)

func getStartSeconds(filename string) (int64, error) {
	var seconds int64
	pcap, err := libpcap.OpenOffline(filename)
	if err != nil {
		return 0, err
	}
	packet, err := pcap.Next()
	if err != nil {
		return 0, err
	}
	if packet != nil {
		seconds = packet.Seconds()
	}
	pcap.Close()
	return seconds, nil
}

// filterOnStartTime returns the list of files that may contain
// packets that occur on are after the provided startTime.
//
// The list of files returned is just a slice of the provided list
// with the start index updated.
func filterOnStartTime(directory string, files []os.FileInfo, startTime int64) []os.FileInfo {

	startIdx := 0

	for idx, file := range files {
		fileStartTime, err := getStartSeconds(path.Join(directory, file.Name()))
		if err != nil {
			continue
		}
		if fileStartTime >= startTime {
			break
		}
		startIdx = idx
	}

	return files[startIdx:]
}

// getFiles returns a list of files in the named directory that start
// with the given prefix. Filenames are sorted per ioutil.ReadDir.
func getFiles(directory string, prefix string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	filtered := make([]os.FileInfo, len(files))
	filteredIdx := 0

	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			filtered[filteredIdx] = file
			filteredIdx++
		}
	}

	return filtered[0:filteredIdx], nil
}

// Dumper sub-program entry point.
func DumperMain(args []string) {

	var directory string
	var prefix string
	var startTime int64
	var duration int64
	var filter string

	log.SetFlags(0)
	log.SetPrefix("dumpy dump: ")

	flagset := flag.NewFlagSet("dumper", flag.ExitOnError)
	flagset.StringVar(&directory, "directory", "", "capture directory")
	flagset.StringVar(&prefix, "prefix", "", "filename prefix")
	flagset.Int64Var(&startTime, "start-time", 0,
		"start time in unix time (seconds)")
	flagset.Int64Var(&duration, "duration", 0, "duration of capture (seconds)")
	flagset.StringVar(&filter, "filter", "", "bpf filter expression")
	flagset.Parse(args)

	if directory == "" || prefix == "" {
		log.Fatalf("-directory and -prefix required")
	}
	if startTime == 0 || duration == 0 {
		log.Fatalf("-start-time and -duration required")
	}

	files, err := getFiles(directory, prefix)
	if err != nil {
		log.Fatal(err)
	}
	files = filterOnStartTime(directory, files, startTime)

	var dumper *libpcap.Dumper

	for _, file := range files {
		log.Printf("opening file %s", file.Name())
		pcap, err := libpcap.OpenOffline(path.Join(directory, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		if filter != "" {
			err := pcap.CompileAndSetFilter(filter)
			if err != nil {
				log.Fatal(err)
			}
		}

		for {
			packet, err := pcap.Next()
			if err != nil {
				log.Printf("warning: %s", err)
			} else if packet == nil {
				break
			} else {
				if packet.Seconds() < startTime {
					continue
				} else if packet.Seconds() > startTime+duration {
					break
				}
				if dumper == nil {
					dumper, err = libpcap.DumpOpen(pcap, "-")
					if err != nil {
						log.Fatalf("failed top open dumper: %s", err)
					}
				}
				dumper.Dump(packet)
			}
		}

		pcap.Close()

	}

	if dumper != nil {
		dumper.Close()
	}
}
