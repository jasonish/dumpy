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

# A non-blocking process wrapper around the dumpy-extracter program.

from __future__ import print_function

import sys
import os
import logging
import subprocess
import time

import tornado.ioloop

import dumpy.util

BLOCK_SIZE = 8192

logger = logging.getLogger("dumpy.extractersubprocess")

def get_command():
    """ Get the command for the dumpy-extract process.  We derive the
    patah to dumpy-extract from the path of the web process. """
    return [os.path.dirname(sys.argv[0]) + "/dumpy-extract"]

class ExtracterSubprocess(object):

    def __init__(self, spool_directory, spool_prefix, options={}):
        self.options = options
        self.spool_directory = spool_directory
        self.spool_prefix = spool_prefix
        self.ioloop = tornado.ioloop.IOLoop.instance()

        self.stdout_handler = None
        self.stderr_handler = None
        self.finished_callback = None

        self.stdout_closed = False
        self.stderr_closed = False

    def add_callbacks(
        self, stdout_callback=None, stderr_callback=None, 
        finished_callback=None):
        self.stdout_handler = stdout_callback
        self.stderr_handler = stderr_callback
        self.finished_callback = finished_callback

    def build_command(self):
        command = get_command()
        command += ["-d", self.spool_directory]
        command += ["-p", self.spool_prefix]
        if self.options["start-time"]:
            command += ["-s", self.options["start-time"]]
        if self.options["end-time"]:
            command += ["-e", self.options["end-time"]]
        if self.options["tzoffset"]:
            command += ["-t", self.options["tzoffset"]]
        if self.options["query"]:
            command += [self.options["query"]]
        return command

    def start(self):
        command = self.build_command()
        logger.info("Running %s", " ".join(command))
        self.child = subprocess.Popen(
            command, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        
        # Put stdout and stderr into non-blocking mode.
        dumpy.util.set_nonblocking(self.child.stdout.fileno())
        dumpy.util.set_nonblocking(self.child.stderr.fileno())
        
        # Start reading stderr right away, we don't do flow control on
        # it.
        self.ioloop.add_handler(
            self.child.stderr.fileno(), self._stderr_handler, self.ioloop.READ)

        self.start_reading()

    def stop(self):
        self.child.terminate()

    def start_reading(self):
        self.ioloop.add_handler(
            self.child.stdout.fileno(), self._stdout_handler, self.ioloop.READ)
        
    def stop_reading(self):
        self.ioloop.remove_handler(self.child.stdout.fileno())
            
    def _stdout_handler(self, fd, events):
        data = os.read(fd, BLOCK_SIZE)
        if data:
            if self.stdout_handler:
                self.stdout_handler(data)
        else:
            self.ioloop.remove_handler(fd)
            self.stdout_closed = True
            self.stdout_handler(None)
            self.check_done()

    def _stderr_handler(self, fd, events):
        data = os.read(fd, BLOCK_SIZE)
        if data:
            if self.stderr_handler:
                self.stderr_handler(data)
        else:
            self.ioloop.remove_handler(fd)
            self.stderr_closed = True
            self.stderr_handler(None)
            self.check_done()

    def check_done(self):
        if self.stdout_closed and self.stderr_closed:
            rc = self.child.poll()
            if rc is None:
                self.ioloop.add_timeout(time.time() + 0.1, self.check_done)
                return
            logger.info("extracter exited with code %d", rc)
            self.finished_callback(rc)
            
