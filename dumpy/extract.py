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

# A program somewhat like tcpslice that processes a directory of pcap
# dump files and extracts the packets matching a timestamp filter as
# well as a bpf filter.

from __future__ import print_function

import sys
import os
import getopt
import glob
import time
import logging
import datetime

# From dumpy.
from dumpy import eventdecoders
from dumpy import timetools
from dumpy import libpcap

logging.basicConfig(
    level=logging.DEBUG, format="%(levelname)s:%(message)s")
logger = logging.getLogger()

def errx(status, msg):
    """ Print an error message then exit. """
    print(msg, file=sys.stderr)
    sys.exit(status)

def printerr(msg):
    """ Print a message to sys.stderr. """
    print(msg, file=sys.stderr)

def get_start_timestamp(filename):
    """ Get the start timestamp for the provided filename.  
    
    This means opening the file and getting the timestamp for the
    first event.
    """
    pcap = libpcap.pcap_open_offline(filename)
    packet = pcap.get_next()
    if packet:
        tv = packet.header.get_tv()
    else:
        tv = None
    pcap.close()
    return tv

def get_files(directory, prefix):
    """ Return a sorted list of files from the provided directory with
    the given prefix. """

    # First get the list of files.
    files = glob.glob("%s/%s*" % (directory, prefix))

    # Return the files sorted by their modified time.
    return sorted(files, key=lambda x: get_start_timestamp(x).secs)

def filter_files(files, filter_start_time, filter_end_time):
    """ This function takes a list of pcap files and filter them based
    on the passed in timestamp filter parameters.

    If a filter end time is set, any files where the first event
    occurs after the filter start time will be discarded.

    If a filter start time is set, all files leading up to the one
    that contains the start time will be discarded.
    """

    def filter_on_end_time(files, filter_end_time):
        filtered = []
        for filename in files:
            if get_start_timestamp(filename) < filter_end_time:
                filtered.append(filename)
            else:
                break
        return filtered

    def filter_on_start_time(files, filter_start_time):
        filtered = []
        for filename in reversed(files):
            filtered.append(filename)
            if get_start_timestamp(filename) < filter_start_time:
                break;
        filtered.reverse()
        return filtered

    if filter_end_time:
        files = filter_on_end_time(files, filter_end_time)

    if filter_start_time:
        files = filter_on_start_time(files, filter_start_time)

    return files

def extract(files, start_time, end_time, pcap_filter, output):
    dumper = None
    for filename in files:
        logger.debug("Processing file %s", filename)
        pcap = libpcap.pcap_open_offline(filename)
        pcap.set_filter(pcap_filter)
        
        while True:
            packet = pcap.get_next()
            if not packet:
                break
            if start_time and packet.header.get_tv() < start_time:
                continue
            if end_time and packet.header.get_tv() >= end_time:
                break
            
            if not dumper:
                dumper = libpcap.PcapDumper(pcap, output)
            dumper.dump(packet)

        pcap.close()
    
    if dumper:
        dumper.close()

def usage(file=sys.stderr):
    print("""
usage: %(progname)s [options] [filter]

options:

  -s <start-timestamp>     Start timestamp
  -e <end-timestamp>       End timestamp
  -d <directory>           Directory to process
  -p <prefix>              Filename prefix for files to process
  -o <filename>            Output filename (- for stdout, default)
  -t <tzoffset>            Timezone offset to use if not provided in timestamp
""" % {"progname": sys.argv[0]}, file=file)

def main():

    opt_directory = None
    opt_prefix = None
    opt_start_time = None
    opt_end_time = None
    opt_output = "-"
    opt_tzoffset = None

    try:
        opts, args = getopt.getopt(
            sys.argv[1:], "d:ho:p:e:s:t:",
            ["help"])
    except getopt.GetoptError as err:
        logger.error("invalid command: %s", err)
        usage(sys.stderr)
        return 1
    for o, a in opts:
        if o in ["-h", "--help"]:
            usage(sys.stdout)
            return 0
        elif o in ["-d"]:
            opt_directory = a
        elif o in ["-p"]:
            opt_prefix = a
        elif o == "-s":
            opt_start_time = a
        elif o == "-e":
            opt_end_time = a
        elif o == "-o":
            opt_output = a
        elif o == "-t":
            opt_tzoffset = a

    if not opt_directory:
        logger.error("required parameter missing: -d")
        usage()
        return 1
    if not opt_prefix:
        logger.error("required parameter missing: -p")
        usage()
        return 1

    if args:
        pcap_filter, reference_time = eventdecoders.decode(" ".join(args))
    else:
        pcap_filter = reference_time = None

    if opt_start_time:
        opt_start_time = timetools.decode(opt_start_time, opt_tzoffset)
        if isinstance(opt_start_time, timetools.TimevalOffset):
            if not reference_time:
                opt_start_time = timetools.Timeval() - opt_start_time
            else:
                opt_start_time = reference_time - opt_start_time
        if not reference_time:
            reference_time = opt_start_time

    if opt_end_time:
        opt_end_time = timetools.decode(opt_end_time, opt_tzoffset)
        if isinstance(opt_end_time, timetools.TimevalOffset):
            opt_end_time = reference_time + opt_end_time

    if opt_start_time:
        logger.debug("Resolved start time to %s" % (
                opt_start_time.to_datetime()))
    if opt_end_time:
        logger.debug("Resolved end time to %s" % (
                opt_end_time.to_datetime()))

    files = get_files(opt_directory, opt_prefix)
    if not files:
        logger.error("no files found in %s", opt_directory)
        return 1

    files = filter_files(files, opt_start_time, opt_end_time)
    if not files:
        logger.error("no packets found within specified time")
        return 1

    extract(files, opt_start_time, opt_end_time, pcap_filter, opt_output)

if __name__ == "__main__":
    sys.exit(main())
