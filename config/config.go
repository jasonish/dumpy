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

package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"gopkg.in/yaml.v1"

	"crypto/md5"
	"golang.org/x/crypto/bcrypt"
)

const (
	DEFAULT_PORT = 7000
	DEFAULT_TLS_KEY_FILENAME = "key.pem"
	DEFAULT_TLS_CERT_FILENAME = "cert.pem"
)

type TlsConfig struct {
	Enabled     bool   `json:"enabled"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

type SpoolConfig struct {
	Name      string `json:"name"`
	Directory string `json:"directory"`
	Prefix    string `json:"prefix"`
	Recursive bool `json:"recursive"`
}

type Config struct {
	filename string
	checksum []byte

	Port     int               `json:"port"`
	Tls      TlsConfig         `json:"tls"`
	Spools   []*SpoolConfig    `json:"spools"`
	Users    map[string]string `json:"users"`
}

func NewConfig(filename string) *Config {
	config := Config{filename: filename}

	// Set some defaults.
	config.Port = DEFAULT_PORT
	config.Tls.Enabled = false
	config.Tls.Certificate = DEFAULT_TLS_CERT_FILENAME
	config.Tls.Key = DEFAULT_TLS_KEY_FILENAME
	config.Users = make(map[string]string)
	config.Spools = make([]*SpoolConfig, 0)

	if filename != "" {
		buf, err := ioutil.ReadFile(filename)
		if err == nil {
			err = yaml.Unmarshal(buf, &config)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	config.checksum = config.Checksum()

	return &config
}

// Get a spool by name.
func (c *Config) GetSpoolByName(name string) *SpoolConfig {
	for _, spool := range c.Spools {
		if spool.Name == name {
			return spool
		}
	}
	return nil
}

// Get the first spool listed (the default).
func (c *Config) GetFirstSpool() *SpoolConfig {
	if len(c.Spools) > 0 {
		return c.Spools[0]
	}
	return nil
}

// Get a spool by name falling back to the default.
func (c *Config) GetSpoolByNameOrDefault(name string) *SpoolConfig {
	spool := c.GetSpoolByName(name)
	if spool != nil {
		return spool
	}
	return c.GetFirstSpool()
}

func (c *Config) Checksum() []byte {
	buf, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("failed to generate config checksum: %s", err)
		return nil
	}
	checksumBuilder := md5.New()
	checksumBuilder.Write(buf)
	checksum := checksumBuilder.Sum(nil)
	return checksum
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *Config) Write() {

	if bytes.Equal(c.checksum, c.Checksum()) {
		return
	}

	buf, err := yaml.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	tmp := c.filename + ".tmp"
	if err = ioutil.WriteFile(tmp, buf, 0600); err != nil {
		log.Fatal(err)
	}
	bak := c.filename + ".bak"
	os.Remove(bak)
	os.Link(c.filename, bak)
	os.Remove(c.filename)
	os.Link(tmp, c.filename)
	os.Remove(tmp)
}

func PasswdCommand(config *Config, args []string) {

	username := args[0]
	password := args[1]

	cryptedPassword, err := bcrypt.GenerateFromPassword(([]byte)(password), 0)
	if err != nil {
		log.Fatalf("failed to encrypt password: %s", err)
	}

	if _, ok := config.Users[username]; ok {
		log.Printf("Updating password for user %s.", username)
	} else {
		log.Printf("Creating new user %s.", username)
	}

	config.Users[username] = string(cryptedPassword)
}

func SetCommand(config *Config, args []string) error {
	if len(args) < 2 {
		return errors.New("dumpy config set: not enough arguments")
	}

	parameter := args[0]
	value := args[1]

	switch parameter {
	case "port":
		port, err := strconv.ParseInt(value, 10, 16)
		if err != nil || (port < 1 || port > 65535) {
			return fmt.Errorf("dumpy config set: invalid port value: %s", value)
		}
		config.Port = (int)(port)
	case "tls.enabled":
		tlsEnabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("dumpy config set: invalid value: %s", value)
		}
		config.Tls.Enabled = tlsEnabled
	case "tls.certificate":
		config.Tls.Certificate = value
	case "tls.key":
		config.Tls.Key = value
	default:
		return fmt.Errorf("dumpy config set: unknown parameter: %s", parameter)
	}

	return nil
}

func ConfigUsage() {
	fmt.Printf(`
Usage: dumpy config [options] <command>

Commands:

    passwd                     Create users and set passwords
    set <name> <value>         Set a configuration value
    show                       Show the configuration
    spool                      Add or remove spool directories

Configuration parameters:

    port                       Port to listen on
    tls.enabled <true|false>   Enable/disable TLS
    tls.certificate <filename> TLS certificate filename
    tls.key <filename>         TLS key filename

`)
}

func ConfigMain(config *Config, args []string) {

	log.SetPrefix("")
	log.SetFlags(0)

	flagset := flag.NewFlagSet("config", flag.ExitOnError)
	flagset.Parse(args)

	args = flagset.Args()
	if len(args) == 0 {
		ConfigUsage()
		return
	}

	switch args[0] {
	case "passwd":
		PasswdCommand(config, args[1:])
	case "set":
		if err := SetCommand(config, args[1:]); err != nil {
			log.Fatal(err)
		}
	case "show":
		buf, err := config.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	case "spool":
		if (&SpoolCommand{config}).Run(args[1:]) != 0 {
			os.Exit(1)
		}
	default:
		ConfigUsage()
		return
	}

	config.Write()
}
