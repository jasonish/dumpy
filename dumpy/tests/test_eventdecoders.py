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

from __future__ import print_function

import unittest

import dumpy.eventdecoders
from dumpy import timetools

class SnortFastEventDecoderTests(unittest.TestCase):
    
    def setUp(self):
        self.snortFastEventDecoder = dumpy.eventdecoders.SnortFastEventDecoder()
    
    def test_decode_tcp_brief(self):
        f = self.snortFastEventDecoder.decode(
            "{TCP} 217.160.51.31:80 -> 172.16.1.11:33189")
        self.assertEquals(
            "proto 6 and ((host 217.160.51.31 and port 80) and "
            "(host 172.16.1.11 and port 33189))", f.filter)
        self.assertEquals(None, f.timestamp)

    def test_decode_tcp_full(self):
        f = self.snortFastEventDecoder.decode(
            "11/15-22:56:29.943914  [**] [1:498:8] INDICATOR-COMPROMISE id check returned root [**] [Classification: Potentially Bad Traffic] [Priority: 2] {TCP} 217.160.51.31:80 -> 172.16.1.11:33189")
        self.assertEquals(
            "proto 6 and ((host 217.160.51.31 and port 80) and "
            "(host 172.16.1.11 and port 33189))", f.filter)
        self.assertEqual(timetools.Timeval(1353041789, 943914), f.timestamp)

    def test_suricata_decode_tcp_full(self):
        f = self.snortFastEventDecoder.decode(
            "11/15/2012-22:56:29.943914  [**] [1:2100498:7] GPL ATTACK_RESPONSE id check returned root [**] [Classification: Potentially Bad Traffic] [Priority: 2] {TCP} 217.160.51.31:80 -> 172.16.1.11:33188")
        self.assertEquals(
            "proto 6 and ((host 217.160.51.31 and port 80) and "
            "(host 172.16.1.11 and port 33188))", f.filter)
        self.assertEqual(timetools.Timeval(1353041789, 943914), f.timestamp)
