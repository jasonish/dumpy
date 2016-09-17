// Copyright (c) 2016 Jason Ish. All rights reserved.
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

package dumper

import (
	"net/http"
	"strconv"
	"log"
	"os"
	"os/exec"
	"bufio"
	"io"
	"fmt"
)

// DumperProxy - proxies the output from
type DumperProxy struct {
	Options  DumperOptions
	Writer   http.ResponseWriter
	Filename string
}

func (dp *DumperProxy) BuildArgs() []string {
	args := []string{"dump"}

	if dp.Options.Directory != "" {
		args = append(args, "-directory")
		args = append(args, dp.Options.Directory)
	}

	if dp.Options.Prefix != "" {
		args = append(args, "-prefix")
		args = append(args, dp.Options.Prefix)
	}

	args = append(args, []string{"-start-time",
		strconv.FormatInt(dp.Options.StartTime, 10)}...)

	if dp.Options.Duration > 0 {
		args = append(args, "-duration")
		args = append(args, strconv.FormatInt(dp.Options.Duration, 10))
	}

	if dp.Options.Filter != "" {
		args = append(args, "-filter")
		args = append(args, dp.Options.Filter)
	}

	if dp.Options.Recursive {
		args = append(args, "-recursive")
	}

	args = append(args, "-verbose")

	return args
}

func (dp *DumperProxy) Run() {

	args := dp.BuildArgs()

	dumper := exec.Command(os.Args[0], args...)

	stderr, err := dumper.StderrPipe()
	if err != nil {
		log.Println(err)
		return
	}

	stdout, err := dumper.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}

	// Handle stderr output in a go routine.
	go func() {
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				break;
			}
			log.Printf("dumpy dump (stderr): %s", line)
		}
	}()

	err = dumper.Start()
	log.Println(err)

	bytesWritten := 0
	buf := make([]byte, 8192)

	for {
		n, err := stdout.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("error reading from stdin: %s", err.Error())
				if bytesWritten == 0 {
					http.Error(dp.Writer, err.Error(), http.StatusInternalServerError)
				}
			}
			break
		}

		if bytesWritten == 0 {
			dp.Writer.Header().Add("content-type", "application/vnd.tcpdump.pcap")
			dp.Writer.Header().Add("content-disposition",
				fmt.Sprintf("attachment; filename=%s", dp.Filename))
		}

		m, err := dp.Writer.Write(buf[0:n])
		if err != nil {
			log.Printf("Failed to write to http client: %v", err.Error())
		}
		if m != n {
			log.Printf("Didn't write all the bytes??")
		}

		bytesWritten += n

	}

	if bytesWritten == 0 {
		http.Error(dp.Writer, "No packets found.", http.StatusNotFound)
	}

	dumper.Wait();

	log.Printf("Wrote %d bytes of pcap data.", bytesWritten)
}