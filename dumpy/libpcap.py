# Copyright (c) 2012 Jason Ish
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
#
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
#
# THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED
# WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
# MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT,
# INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
# (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
# SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
# HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
# STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING
# IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

# ISO like timestamp parsing functions so we don't have to rely on any
# third party modules.

from __future__ import print_function

import sys
import os
import collections
import ctypes
import ctypes.util

from dumpy import timetools

# Initialize of the underlying libpcap and libc libraries using ctypes.
PCAP_ERRBUF_SIZE = 256
libc = ctypes.cdll.LoadLibrary(ctypes.util.find_library("c"))
libpcap = ctypes.cdll.LoadLibrary(ctypes.util.find_library("pcap"))
libpcap.pcap_geterr.restype = ctypes.c_char_p
pcap_errbuf = ctypes.create_string_buffer(PCAP_ERRBUF_SIZE)

# Named tuple to represent a packet returned from libpcap.
Packet = collections.namedtuple("Packet", ["header", "data"])

class PcapError(Exception):
    def __init__(self, value):
        super(PcapError, self).__init__("libpcap error: %s" % (value))

class pcap_pkthdr(ctypes.Structure):
    """ Python/Ctypes wrapper around the pcap packet header
    struct. """
    _fields_ = [
        ("ts_sec", ctypes.c_ulong),
        ("ts_usec", ctypes.c_ulong),
        ("caplen", ctypes.c_ulong),
        ("pktlen", ctypes.c_ulong)
        ]

    def get_tv(self):
        return timetools.Timeval(self.ts_sec, self.ts_usec)

class Pcap(object):

    def __init__(self, pcap):
        self.pcap = pcap
        self.pkt_header = ctypes.pointer(pcap_pkthdr())
        self.pkt_data = ctypes.c_void_p()

    def close(self):
        libpcap.pcap_close(self.pcap)

    def set_filter(self, pcap_filter):
        bpf_program = ctypes.c_void_p()
        r = libpcap.pcap_compile(
            self.pcap, ctypes.byref(bpf_program), pcap_filter, 1, 0)
        if r != 0:
            raise PcapError("failed to compile filter: %s: %s" % (
                pcap_filter, self.get_error()))
        if libpcap.pcap_setfilter(self.pcap, ctypes.byref(bpf_program)) != 0:
            raise PcapError("failed to set filter: %s" % (
                    self.get_error()))
        
    def get_error(self):
        return libpcap.pcap_geterr(self.pcap)

    def get_next(self):
        rc = libpcap.pcap_next_ex(
            self.pcap, ctypes.byref(self.pkt_header), 
            ctypes.byref(self.pkt_data))
        if rc == 1:
            return Packet(self.pkt_header.contents, self.pkt_data)
        elif rc in [-2, 0]:
            return None
        else:
            raise PcapError(self.get_error())

class PcapDumper(object):
    """ Wrapper around pcap dumpers. """

    def __init__(self, pcap, filename):
        if filename == "-":
            # We should be able to pass "-" directory to
            # pcap_dump_open, but I suspect ctypes is doing something
            # that is making it fail, so do this hack to write to
            # stdout.
            self.dumper = libpcap.pcap_dump_fopen(
                pcap.pcap, libc.fdopen(libc.dup(sys.stdout.fileno()), "a"))
        else:
            self.dumper = libpcap.pcap_dump_open(pcap.pcap, filename)

    def dump(self, packet):
        libpcap.pcap_dump(self.dumper, ctypes.byref(packet.header), packet.data)
        libpcap.pcap_dump_flush(self.dumper)

    def close(self):
        libpcap.pcap_dump_close(self.dumper)

def pcap_open_offline(filename):
    pcap = libpcap.pcap_open_offline(filename, pcap_errbuf)
    if not pcap:
        raise PcapError("pcap_open_offline failed: %s" % (
                pcap.errbuf.value))
    return Pcap(pcap)
