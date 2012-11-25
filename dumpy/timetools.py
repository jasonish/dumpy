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

import datetime
import calendar
import time
import time
import re
import collections

class InvalidTimestamp(Exception):
    pass

class InvalidTimezoneOffset(Exception):
    pass

class Timeval(object):
    """ A class representing a unix timeval. """
    
    def __init__(self, second=None, microsecond=None):
        if second is None:
            second = time.time()
        if isinstance(second, float):
            self.secs = int(second * 1000000 / 1000000)
            self.usecs = int(second * 1000000 % 1000000)
        else:
            self.secs = second
            self.usecs = 0 if microsecond is None else microsecond

    def to_datetime(self):
        return datetime.datetime.fromtimestamp(self.__float__())

    def to_utcdatetime(self):
        return datetime.datetime.utcfromtimestamp(self.__float__())

    def __cmp__(self, other):
        if self.secs < other.secs:
            return -1
        elif self.secs > other.secs:
            return 1
        elif self.usecs < other.usecs:
            return -1
        elif self.usecs > other.usecs:
            return 1
        else:
            return 0

    def __getitem__(self, key):
        if key == 0:
            return self.secs
        elif key == 1:
            return self.usecs
        else:
            raise IndexError("invalid index")

    def __float__(self):
        return self.secs + (float(self.usecs) / 1000000)

    def __repr__(self):
        return "(%s, %s)" % (str(self.secs), str(self.usecs))

    def __add__(self, other):
        secs = self.secs + other.secs
        usecs = self.usecs + other.usecs
        if usecs >= 1000000:
            secs += 1
            usecs -= 1000000
        return Timeval(secs, usecs)

    def __sub__(self, other):
        secs = self.secs - other.secs
        usecs = self.usecs - other.usecs
        if usecs < 0:
            secs -= 1
            usecs += 1000000
        return Timeval(secs, usecs)

class TimevalOffset(Timeval):
    """ A class to represent a time offset. """

def tzoffset_to_timedelta(tzoffset):
    """ Convert a timezone offset string into timedelta object. """
    if tzoffset is None:
        return None
    elif tzoffset == "Z":
        return datetime.timedelta()
    else:
        m = re.search("([+-])(\d{2})(\d{2})?", tzoffset)
        if not m:
            raise InvalidTimezoneOffset(tzoffset)
        sign = m.group(1)
        hours = int(m.group(2))
        mins = int(m.group(3)) if m.group(3) else 0
        delta = datetime.timedelta(hours=hours, minutes=mins)
        if sign == "-":
            return datetime.timedelta() - delta
        else:
            return delta

def is_relative(buf):
    """ Return true if the string looks like a relative timestamp. """
    if re.match("^\d+(\s+)?([SsMmHhDd].*)", buf):
        return True
    return False

def decode_relative_timestamp(timestamp):
    m = re.match("(\d+)\s*(\w+)", timestamp)
    if not m:
        raise InvalidTimestamp("Invalid relative timestamp: %s" % (timestamp))

    value = m.group(1)
    interval = m.group(2)

    interval_multipliers = {
        "seconds": 1,
        "minutes": 60,
        "hours": 60 * 60,
        "days": 24 * 60 * 60,
        }

    def get_interval(interval):
        """ Derive the interval from possible abbreviations.  Doesn't
        handle ambiguities, but we don't have any. """
        for i in interval_multipliers:
            if i[0:len(interval)] == interval:
                return i
        raise InvalidTimestamp(
            "Bad relative interval: %s" % (interval))
        
    seconds = int(value) * interval_multipliers[get_interval(interval)]

    return TimevalOffset(seconds, 0)

# A regex that pulls out the parts of an ISO timestamp.
timestamp_pattern = re.compile(
    ("^(\d{4})-?(\d{2})?-?(\d{2})?"
     "[Tt\s]?"
     "(\d\d?)?:?(\d{2})?:?(\d{2})?"
     "(\.\d+)?"
     "([+-]\d+|[Zz])?$"))

def decode_timestamp(buf, default_tzoffset=None):
    m = timestamp_pattern.match(buf)
    if not m:
        raise InvalidTimestamp(buf)
    tzoffset = m.group(8) if m.group(8) else default_tzoffset
    dt = datetime.datetime(
        year=int(m.group(1)),
        month=int(m.group(2)) if m.group(2) else 1,
        day=int(m.group(3)) if m.group(3) else 1,
        hour=int(m.group(4)) if m.group(4) else 0,
        minute=int(m.group(5)) if m.group(5) else 0,
        second=int(m.group(6)) if m.group(6) else 0,
        microsecond=(0 if m.group(7) is None 
                     else int(float(m.group(7)) * 1000000)),
        )
    if tzoffset:
        dt -= tzoffset_to_timedelta(tzoffset)
        return Timeval(
            calendar.timegm(dt.timetuple()), dt.microsecond)
    else:
        return Timeval(time.mktime(dt.timetuple()))

def decode(timestamp, default_tzoffset=None):
    if is_relative(timestamp):
        return decode_relative_timestamp(timestamp)
    else:
        return decode_timestamp(timestamp, default_tzoffset)
