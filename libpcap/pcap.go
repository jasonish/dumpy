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

type Pcap struct {
	pcap *C.pcap_t
}

type PcapPktHdr struct {
	pcap_pkthdr *C.struct_pcap_pkthdr
}

func PcapOpenOffline(filename string) (*Pcap, error) {
	var pcap Pcap
	var errbuf *C.char = (*C.char)(C.calloc(1024, 1))
	pcap.pcap = C.pcap_open_offline(C.CString(filename), errbuf)
	if pcap.pcap == nil {
		return nil, errors.New(C.GoString(errbuf))
	}
	return &pcap, nil
}

func (p *Pcap) GetErr() string {
	return C.GoString(C.pcap_geterr(p.pcap))
}

func (p *Pcap) SetFilter(filter string) error {
	var bpf C.struct_bpf_program
	if C.pcap_compile(p.pcap, &bpf, C.CString(filter), 1, 0) != 0 {
		return errors.New(p.GetErr())
	}

	if C.pcap_setfilter(p.pcap, &bpf) != 0 {
		return errors.New(p.GetErr())
	}

	return nil
}

func (p *Pcap) Next() (*PcapPktHdr, unsafe.Pointer, error) {
	var hdr PcapPktHdr
	var pkt *C.u_char
	rc := C.pcap_next_ex(p.pcap, &hdr.pcap_pkthdr, &pkt)
	if rc == 1 {
		return &hdr, unsafe.Pointer(pkt), nil
	} else if rc == 0 || rc == -2 {
		return nil, nil, nil
	}
	return nil, nil, errors.New(p.GetErr())
}
