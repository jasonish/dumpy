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
import os
import os.path
import time
import logging
import subprocess
import base64
import uuid

import tornado.ioloop
import tornado.web
import tornado.httpserver

import dumpy.auth
import dumpy.util
import dumpy.extractsubprocess
import dumpy.eventdecoders

logger = logging.getLogger("dumpy.web")

# Our application home directory.
app_home = os.path.abspath(os.path.dirname(__file__)) + "/.."

class DumpyRequestHandler(tornado.web.RequestHandler):

    session = {}

    def get_authorization(self):
        auth_header = self.request.headers.get("authorization")
        if not auth_header:
            return None
        auth_type, auth_info = auth_header.split(" ", 2)
        username, password = base64.decodestring(auth_info).split(":", 2)
        return (username, password)

    default_template_args = {
        "query": "",
        "start-time": "",
        "end-time": "",
        "warning": None,
        "error": None,
        }

    def render(self, template_name, **kwargs):
        for key in self.default_template_args:
            if not key in kwargs:
                kwargs[key] = self.default_template_args[key]
        for key in kwargs.keys():
            key0 = key.replace("-", "_")
            if key0 != key:
                kwargs[key0] = kwargs[key]
                del(kwargs[key])
        tornado.web.RequestHandler.render(self, template_name, **kwargs)

class FetchRequestHandler(DumpyRequestHandler):

    def initialize(self):
        self.config = self.settings["config"]
        self.bytes = 0

        # Buffer for data received in stderr.
        self.stderr_buffer = []

    @tornado.web.asynchronous
    @dumpy.auth.require
    def get(self):
        logger.debug(self.request.headers["accept"])
        self.args = {
            "query": self.get_argument("query", None),
            "start-time": self.get_argument("start-time", None),
            "end-time": self.get_argument("end-time", None),
            "tzoffset": self.get_argument("tzoffset", None),
            }
        self.extracter = dumpy.extractsubprocess.ExtracterSubprocess(
            self.config.get("spool", "directory"),
            self.config.get("spool", "prefix"),
            self.args)
        self.extracter.add_callbacks(
            self.stdout_cb, self.stderr_cb, self.exit_cb)

        try:
            self.extracter.start()
        except Exception as err:
            logger.error("Failed to execute \"%s\": %s" % (
                    " ".join(self.extracter.build_command()), err))
            raise

    def post(self):
        """ The get handler can handle post as well. """
        return self.get()

    def flush_done(self):
        # Flush is done, start reading again.
        self.extracter.start_reading()

    def stdout_cb(self, buf):
        if buf:
            if self.bytes == 0:
                # Set headers on the first bytes seen.
                self.set_header("content-type",
                                "application/vnd.tcpdump.pcap")
                self.set_header("Content-Disposition",
                                "attachment; filename=dumpy.pcap");
            self.bytes += len(buf)
            self.write(buf)
            self.flush(callback=self.flush_done)
            self.extracter.stop_reading()

    def stderr_cb(self, buf):
        if buf:
            self.stderr_buffer.append(buf)
            logger.error("dumpy-extract:stderr: %s", buf.rstrip())

    def exit_cb(self, exit_code):
        if exit_code == 0 and self.bytes == 0:
            return self.send_response(
                status=404,
                warning="No packets matched filter and/or time range.")
        elif exit_code != 0 and self.bytes == 0:
            return self.send_response(
                400,
                error="".join(self.stderr_buffer))
        elif exit_code != 0:
            logger.warning("dumpy-extract exited with error code %d "
                           "unable to report to client as data already sent",
                           exit_code)
        self.finish()

    def send_response(self, status=None, error=None, warning=None):
        if "html" in self.request.headers["accept"]:
            return self.render("index.html", error=error, warning=warning)
        else:
            self.set_header("content-type", "text/plain")
            if status:
                self.set_status(status)
            if error:
                self.write(error)
            if warning:
                self.write(warning)
            self.write("\n")
            self.finish()

    def on_connection_close(self):
        """ Called by tornado when the client closes the connection. """
        self.extracter.stop()
        
class IndexHandler(DumpyRequestHandler):
    
    @dumpy.auth.require
    def get(self):
        return self.render("index.html")

tornado_handlers = [
    # / - the main entry
    (r"/", IndexHandler),

    # Static resources.
    (r"/static/(.*)", tornado.web.StaticFileHandler, {
            "path": "./static"}),

    # The fetch command.
    (r"/fetch", FetchRequestHandler),
]

def run(config):

    package_path = os.path.abspath(os.path.dirname(__file__))
    app_install_path = os.path.abspath(os.path.dirname(__file__) + "/..")

    # Initialize the authenticator.
    if os.path.exists("./etc/users"):
        dumpy.auth.authenticator.load_users("./etc/users")
    else:
        # Disable authentication.
        logger.warn("AUTHENTICATION DISABLED: users files does not exist.")
        dumpy.auth.authenticator = None

    cookie_secret = open(app_install_path + "/etc/cookie-secret").read()
    settings = dict(
        app_install_prefix = app_install_path,
        template_path = package_path + "/templates",
        static_path = package_path + "/static",
        cookie_secret = cookie_secret,
        config = config,
        debug = True,
        )

    if config.has_option("http", "ssl") and config.getboolean("http", "ssl"):
        sslkey = config.get("http", "ssl-key")
        sslcert = config.get("http", "ssl-cert")
        if not os.path.exists(sslkey):
            logger.error("SSL key file %s does not exist.", sslkey)
            return 1
        if not os.path.exists(sslcert):
            logger.error("SSL certificate file %s does not exist", sslcert)
            return 1
        ssl_options={"certfile": sslcert, "keyfile": sslkey}
    else:
        ssl_options=None

    http_port = config.getint("http", "port")
    http_addr = config.get("http", "addr")
    logger.info("Starting on %s://%s:%d" % (
            "https" if ssl_options else "http", http_addr, http_port))
    application = tornado.web.Application(tornado_handlers, **settings)
    http_server = tornado.httpserver.HTTPServer(
        application, ssl_options=ssl_options)
    http_server.listen(http_port)
    tornado.ioloop.IOLoop.instance().start()
