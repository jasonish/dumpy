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

package config

import (
	"log"
	"flag"
)

type SpoolAddCommand struct {
	config *Config
	flags *flag.FlagSet
}

func (c *SpoolAddCommand) Usage() {
	log.Println("usage: dumpy config spool add [OPTIONS]")
	log.Println()
	c.flags.PrintDefaults()
}

func (c *SpoolAddCommand) UsageWithError(error string) {
	log.Println("error:", error)
	c.Usage()
}

func (c *SpoolAddCommand) Run(args []string) int {
	var name string
	var directory string
	var prefix string
	recursive := false

	c.flags = flag.NewFlagSet("dumpy config spool", flag.ExitOnError)
	c.flags.StringVar(&name, "name", "", "Spool name (optional)")
	c.flags.StringVar(&directory, "directory", "", "Spool directory")
	c.flags.StringVar(&prefix, "prefix", "", "Filename prefix")
	c.flags.BoolVar(&recursive, "recursive", false, "Recursive")

	if len(args) == 0 {
		c.Usage()
		return 1
	}

	c.flags.Parse(args)

	if directory == "" {
		c.UsageWithError("a directory is required")
		return 1
	}
	if prefix == "" {
		c.UsageWithError("a prefix is required")
		return 1
	}

	if name == "" {
		name = directory
	}

	if spool := c.config.GetSpoolByName(name); spool != nil {
		log.Printf("error: spool %s already exists", name)
		return 1
	}

	spool := SpoolConfig{
		Name:      name,
		Directory: directory,
		Prefix:    prefix,
		Recursive: recursive,
	}

	c.config.Spools = append(c.config.Spools, &spool)

	return 0
}

type SpoolCommand struct {
	config *Config
}

func (c *SpoolCommand) Usage() {
	log.Printf(`usage: dumpy config spool add
   or: dumpy config spool remove
`)
}

func (c *SpoolCommand) Run(args []string) int {

	flagset := flag.NewFlagSet("dumpy config spool", flag.ExitOnError)
	showUsage := flagset.Bool("h", false, "usage")
	flagset.Usage = c.Usage
	flagset.Parse(args)
	if *showUsage || flagset.NArg() == 0 {
		c.Usage()
		return 1
	}

	command := flagset.Args()[0]
	switch command {
	case "add":
		return (&SpoolAddCommand{config: c.config}).Run(flagset.Args()[1:])
	case "remove":
		return c.Remove(flagset.Args()[1:])
	default:
		log.Printf("dumpy config spool: unknown sub-command: %s", command)
		return 1
	}
}

func (c *SpoolCommand) Remove(args []string) int {
	if len(args) < 1 {
		log.Println("usage: dumpy config spool remove <name>")
		return 1
	}
	spoolName := args[0]

	spool := c.config.GetSpoolByName(spoolName)
	if spool == nil {
		log.Printf("error: no spool named %s", spoolName)
		return 1
	}

	for idx, spool := range c.config.Spools {
		if spool.Name == spoolName {
			c.config.Spools = append(c.config.Spools[:idx],
				c.config.Spools[idx+1:]...)
			break
		}
	}

	return 0
}
