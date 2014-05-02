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

// Package libpcap is a minimal libpcap wrapper for the purposes of
// dumpy.
package libpcap

// #cgo LDFLAGS: -lpcap
// #include <stdlib.h>
// #include <stdio.h>
// #include <pcap.h>
import "C"

import (
	"errors"
	"unsafe"
)

type Packet struct {
	header *C.struct_pcap_pkthdr
	data   *C.u_char
}

func (p *Packet) Seconds() int64 {
	return (int64)(p.header.ts.tv_sec)
}

type Pcap struct {
	pcap *C.pcap_t
}

func OpenOffline(filename string) (*Pcap, error) {
	var pcap Pcap
	var errbuf *C.char = (*C.char)(C.calloc(1024, 1))
	pcap.pcap = C.pcap_open_offline(C.CString(filename), errbuf)
	if pcap.pcap == nil {
		return nil, errors.New(C.GoString(errbuf))
	}
	return &pcap, nil
}

func (p *Pcap) Close() {
	C.pcap_close(p.pcap)
}

func (p *Pcap) GetErr() string {
	return C.GoString(C.pcap_geterr(p.pcap))
}

func (p *Pcap) CompileAndSetFilter(filter string) error {
	var bpf C.struct_bpf_program
	if C.pcap_compile(p.pcap, &bpf, C.CString(filter), 1, 0) != 0 {
		return errors.New(p.GetErr())
	}

	if C.pcap_setfilter(p.pcap, &bpf) != 0 {
		return errors.New(p.GetErr())
	}

	return nil
}

func (p *Pcap) Next() (*Packet, error) {
	var packet Packet
	rc := C.pcap_next_ex(p.pcap, &packet.header, &packet.data)
	if rc == 1 {
		return &packet, nil
	} else if rc == 0 || rc == -2 {
		return nil, nil
	}
	return nil, errors.New(p.GetErr())
}

type Dumper struct {
	dumper *C.pcap_dumper_t
}

func DumpOpen(pcap *Pcap, filename string) (*Dumper, error) {
	var dumper Dumper
	dumper.dumper = C.pcap_dump_open(pcap.pcap, C.CString(filename))
	if dumper.dumper == nil {
		return nil, errors.New(pcap.GetErr())
	}
	return &dumper, nil
}

func (d *Dumper) Close() {
	C.pcap_dump_close(d.dumper)
}

func (d *Dumper) Dump(packet *Packet) {
	C.pcap_dump((*C.u_char)(unsafe.Pointer(d.dumper)), packet.header,
		packet.data)
}
