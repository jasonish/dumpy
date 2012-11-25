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

import re
import collections
import datetime
import time
import calendar

from dumpy import timetools
from dumpy import util

Filter = collections.namedtuple("Filter", ("filter", "timestamp"))

def guess_year(parsed_time):
    now = datetime.datetime.now()
    if datetime.datetime(*(now.year,) + parsed_time[1:6]) <= now:
        return now.year
    else:
        return now.year - 1

class SnortFastEventDecoder(object):
    """ A parser for Snort fast style logs. """

    def __init__(self):
        self.filter_pattern = re.compile(
            "{(\d+|\w+)}\s([\d\.]+):?(\d+)?\s..\s([\d\.]+):?(\d+)?")

        self.timestamp_pattern = re.compile(
            "^(\d\d)\/(\d\d)(?:\/)?(\d{4})?-(\d\d):(\d\d):(\d\d).(\d+)")

    def decode_event(self, event):
        m = self.filter_pattern.search(event)
        if not m:
            return None
        proto = util.getprotobyname(m.group(1))
        src_addr = m.group(2)
        src_port = m.group(3)
        dst_addr = m.group(4)
        dst_port = m.group(5)

        if src_port and src_port != "0":
            src_filter = "(host %s and port %s)" % (
                src_addr, src_port)
        else:
            src_filter = "(host %s)" % (src_addr)

        if dst_port and dst_port != "0":
            dst_filter = "(host %s and port %s)" % (
                dst_addr, dst_port)
        else:
            dst_filter = "(host %s)" % (dst_addr)

        return "proto %s and (%s and %s)" % (
            proto, src_filter, dst_filter)
    
    def parse_timestamp(self, event):
        """ Parse the timestamp into a sequence of the form [yyyy, mm,
        dd, hh, mm, ss, us]. """
        m = self.timestamp_pattern.match(event)
        if m:
            return [int(x) if x else x for x in m.group(3, 1, 2, 4, 5, 6, 7)]
        else:
            return None

    def decode_timestamp(self, event):
        timestamp = self.parse_timestamp(event)
        if timestamp:
            if timestamp[0] is None:
                timestamp[0] = guess_year(tuple(timestamp))
            return (
                timetools.Timeval(
                    int(time.mktime(datetime.datetime(*timestamp).timetuple())),
                    timestamp[-1]))
        return None

    def decode(self, event):
        pcap_filter = self.decode_event(event)
        if not pcap_filter:
            return None
        else:
            return Filter(pcap_filter, self.decode_timestamp(event))

decoders = [
    SnortFastEventDecoder(),
]

def decode(event):
    for decoder in decoders:
        query = decoder.decode(event)
        if query:
            return query
    # If we get here, assume it was a pcap filter to start with.
    return Filter(event, None)
