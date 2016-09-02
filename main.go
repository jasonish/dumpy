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
	"github.com/jasonish/dumpy/dumper"
	"flag"
	"fmt"
	"os"
	"github.com/jasonish/dumpy/config"
	"log"
)

// Global logger.
var logger = NewLogger("")

func Usage() {
	fmt.Fprintf(os.Stderr, `
Usage: dumpy [options] <command>

Options:
    -config <file>       Path to the configuration file

Commands:
    start                Start the server
    version              Display version and exit
    config               Configuration tool
    dump                 Command to process pcap files
    generate-cert        Generate a self signed TLS certificate

`)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	var configFilename string

	flag.Usage = Usage
	flag.StringVar(&configFilename, "config", "dumpy.yaml", "config file")
	flag.Parse()

	if len(flag.Args()) < 1 {
		Usage()
		os.Exit(1)
	} else {
		switch flag.Args()[0] {
		case "version":
			fmt.Println(VERSION)
		case "dump":
			dumper.DumperMain(os.Args[2:])
		case "config":
			config.ConfigMain(config.NewConfig(configFilename), os.Args[2:])
		case "start":
			log.Println("Starting server...")
			StartServer(config.NewConfig(configFilename))
		case "generate-cert":
			GenerateCertMain(os.Args[2:])
		default:
			log.Println("Bad command:", flag.Args()[0])
		}
	}

}
