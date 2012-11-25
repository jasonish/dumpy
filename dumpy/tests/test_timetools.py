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

import sys
import unittest
import calendar
import commands
import time
import datetime

from dumpy import timetools

my_tzoffset = commands.getoutput("date +%z")

class IsRelativeTests(unittest.TestCase):

    def test_minute(self):
        self.assertTrue(timetools.is_relative("1m"))

    def test_hour(self):
        self.assertTrue(timetools.is_relative("2 hours"))

    def test_invalid(self):
        self.assertFalse(timetools.is_relative("2012"))
        self.assertFalse(timetools.is_relative("2012-01-01T12:01:00Z"))

class DecodeRelativeTests(unittest.TestCase):

    def test_bad_relative(self):
        self.assertRaises(timetools.InvalidTimestamp,
                          timetools.decode_relative_timestamp, "bad")

    def check_interval_variations(self, interval, expected):
        """ Runs through all variations of the interval spellings to
        make sure they are matched properly as relative time
        intervals. 

        Currently with interals seconds/minutes/hours/days there are
        no ambiguities with single letter abbreviations.
        """
        for i in range(len(interval)):
            self.assertEquals(
                expected, timetools.decode("1 %s" % (interval[0:i+1])))

    def test_interval_variations(self):
        self.check_interval_variations("seconds", timetools.TimevalOffset(1))
        self.check_interval_variations("minutes", timetools.TimevalOffset(60))
        self.check_interval_variations("hours", timetools.TimevalOffset(60*60))
        self.check_interval_variations(
            "days", timetools.TimevalOffset(60*60*24))

    def test_seconds(self):
        self.assertEquals(
            timetools.TimevalOffset(1),
            timetools.decode("1s"))

    def test_minutes(self):
        self.assertEquals(
            timetools.TimevalOffset(3600),
            timetools.decode("60m"))

    def test_hours(self):
        self.assertEquals(
            timetools.TimevalOffset(7200),
            timetools.decode("2h"))

    def test_days(self):
        self.assertEquals(
            timetools.TimevalOffset(86400*3),
            timetools.decode("3d"))

class DecodeTimestampTests(unittest.TestCase):

    expected_timeval = timetools.Timeval(1352693563, 0)

    def test_without_tz(self):
        """ 
        test_without_tz

        Timestamps specified without a timezone will be treated as
        local time. """
        self.assertEquals(
            timetools.decode("2012-11-11T22:12:43"),
            timetools.decode("2012-11-11T22:12:43%s" % (my_tzoffset)))

    def test_utc(self):
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43Z"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43", "Z"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43-0000"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43-00"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43+0000"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T04:12:43+00"))

    def test_america_regina(self):
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("20121111T221243-0600"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("20121111T221243", "-0600"))

    def test_australia_sydney(self):
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T15:12:43+1100"))
        self.assertEquals(
            self.expected_timeval,
            timetools.decode("2012-11-12T15:12:43", "+1100"))

    def test_year(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1).timetuple()), 0),
            timetools.decode("2012", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(datetime.datetime(2012, 1, 1).timetuple())),
            timetools.decode("2012"))

    def test_year_month(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1).timetuple()), 0),
            timetools.decode("2012-01", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(datetime.datetime(2012, 1, 1).timetuple())),
            timetools.decode("2012-01"))

    def test_year_month_day(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1).timetuple()), 0),
            timetools.decode("2012-01-01", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(datetime.datetime(2012, 1, 1).timetuple())),
            timetools.decode("2012-01-01"))

    def test_year_month_day_hour(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1, 12).timetuple()), 0),
            timetools.decode("2012-01-01T12", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(datetime.datetime(2012, 1, 1, 12).timetuple())),
            timetools.decode("2012-01-01T12"))

    def test_year_month_day_hour_minute(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1, 12, 1).timetuple()), 0),
            timetools.decode("2012-01-01T12:01", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(datetime.datetime(2012, 1, 1, 12, 1).timetuple())),
            timetools.decode("2012-01-01T12:01"))

    def test_year_month_day_hour_minute_second(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1, 12, 1, 2).timetuple()), 0),
            timetools.decode("2012-01-01T12:01:02", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(
                    datetime.datetime(2012, 1, 1, 12, 1, 2).timetuple())),
            timetools.decode("2012-01-01T12:01:02"))

    def test_year_month_day_hour_minute_second_microsecond(self):
        self.assertEquals(
            timetools.Timeval(
                calendar.timegm(
                    datetime.datetime(2012, 1, 1, 12, 1, 2).timetuple()),
                1),
            timetools.decode("2012-01-01T12:01:02.000001", "Z"))
        self.assertEquals(
            timetools.Timeval(
                time.mktime(
                    datetime.datetime(2012, 1, 1, 12, 1, 2).timetuple()),
                1),
            timetools.decode("2012-01-01T12:01:02.000001"))

class TzOffsetToTimedeltaTests(unittest.TestCase):

    def test_invalid(self):
        self.assertRaises(
            timetools.InvalidTimezoneOffset,
            timetools.tzoffset_to_timedelta, "0600")

    def test_utc(self):
        self.assertEquals(
            datetime.timedelta(seconds=0),
            timetools.tzoffset_to_timedelta("Z"))
        self.assertEquals(
            datetime.timedelta(seconds=0),
            timetools.tzoffset_to_timedelta("-0000"))
        self.assertEquals(
            datetime.timedelta(seconds=0),
            timetools.tzoffset_to_timedelta("+0000"))

    def test_minus0600(self):
        self.assertEquals(
            datetime.timedelta() - datetime.timedelta(hours=6),
            timetools.tzoffset_to_timedelta("-0600"))

    def test_plus1200(self):
        self.assertEquals(
            datetime.timedelta(hours=12),
            timetools.tzoffset_to_timedelta("+1200"))
