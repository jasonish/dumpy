// Copyright (c) 2014-2016 Jason Ish. All rights reserved.
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

package cmd

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sort"

	"github.com/jasonish/dumpy/libpcap"
)

var verbose bool = false

// PcapFilenames - type and interface functions to allow filename sorting.
type PcapFilenames []string

func (pf PcapFilenames) Len() int {
	return len(pf)
}

func (pf PcapFilenames) Swap(i, j int) {
	pf[i], pf[j] = pf[j], pf[i]
}

func (pf PcapFilenames) Less(i, j int) bool {
	log.Println("Comparing", pf[i], pf[j])
	iTime, err := getStartSeconds(pf[i])
	if err != nil {
		log.Fatal(err)
	}
	jTime, err := getStartSeconds(pf[j])
	if err != nil {
		log.Fatal(err)
	}
	return iTime < jTime
}

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
func filterOnStartTime(filenames []string, startTime int64) []string {

	startIdx := 0

	for idx, filename := range filenames {
		fileStartTime, err := getStartSeconds(filename)
		if err != nil {
			log.Println("Failed to get start time for file ", filename)
		}
		if fileStartTime >= startTime {
			break
		}
		startIdx = idx
	}

	return filenames[startIdx:]
}

// getFiles returns a list of files in the named directory that start
// with the given prefix. Filenames are sorted per ioutil.ReadDir.
func findFiles(directory string, prefix string, recursive bool) ([]string, error) {

	var pcap_filenames []string

	if recursive {
		err := filepath.Walk(directory,
			func(pathname string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if fileInfo.IsDir() {
					return nil
				}

				if strings.HasPrefix(path.Base(pathname), prefix) {
					pcap_filenames = append(pcap_filenames, pathname)
				}

				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		files, err := ioutil.ReadDir(directory)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) {
				filename := path.Join(directory, file.Name())
				pcap_filenames = append(pcap_filenames, filename)
			}
		}

	}

	sort.Sort(PcapFilenames(pcap_filenames))

	return pcap_filenames, nil
}

// Dumper sub-program entry point.
func DumperMain(args []string) {

	var directory string
	var recursive bool
	var prefix string
	var startTime int64
	var duration int64
	var filter string

	log.SetFlags(0)
	log.SetPrefix("dumpy dump: ")

	flagset := flag.NewFlagSet("dumper", flag.ExitOnError)
	flagset.StringVar(&directory, "directory", "", "capture directory")
	flagset.BoolVar(&recursive, "recursive", false, "process directory recursively")
	flagset.StringVar(&prefix, "prefix", "", "filename prefix")
	flagset.Int64Var(&startTime, "start-time", 0,
		"start time in unix time (seconds)")
	flagset.Int64Var(&duration, "duration", 0, "duration of capture (seconds)")
	flagset.StringVar(&filter, "filter", "", "bpf filter expression")
	flagset.BoolVar(&verbose, "verbose", false, "be more verbose")
	flagset.Parse(args)

	if directory == "" || prefix == "" {
		log.Fatalf("-directory and -prefix required")
	}
	if startTime == 0 || duration == 0 {
		log.Fatalf("-start-time and -duration required")
	}

	filenames, err := findFiles(directory, prefix, recursive)
	if err != nil {
		log.Fatal(err)
	}
	filenames = filterOnStartTime(filenames, startTime)

	var dumper *libpcap.Dumper

	for _, file := range filenames {
		if verbose {
			log.Printf("opening file %s", file)
		}
		pcap, err := libpcap.OpenOffline(file)
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
				} else if packet.Seconds() > startTime + duration {
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
