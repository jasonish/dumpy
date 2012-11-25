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

import os
import logging

try:
    import bcrypt
    has_bcrypt = True
except:
    has_bcrypt = False

logger = logging.getLogger("dumpy.auth")

class NoHasherAvailableError(Exception):
    pass

class BcryptHasher(object):

    prefix = "bcrypt:"

    def hash_password(self, password):
        hash = bcrypt.hashpw(password, bcrypt.gensalt())
        return "%s%s" % (self.prefix, hash)

    def check_password(self, password, hash_in):
        if not hash_in.startswith(self.prefix):
            return False
        method, hash = hash_in.split(":", 2)
        hash_check = bcrypt.hashpw(password, hash)
        return hash == hash_check

class Authenticator(object):

    def __init__(self, users_file=None):
        self.users = {}
        self.users_filename = None
        self.users_filename_last_loaded = 0

        if users_file:
            self.load_users(users_file)

    def load_users(self, filename=None, fileobj=None):
        """ Load users from the provided filename or fileobj,
        replacing the current dict of users. """
        self.users_filename = filename
        users = {}
        if not fileobj:
            self.users_filename_last_loaded = os.path.getmtime(
                self.users_filename)
            fileobj = open(filename)
            
        for line in fileobj:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            username, hash = line.split(":", 1)
            users[username] = hash
        self.users = users

    def reload_users(self):
        if self.users_filename and os.path.exists(self.users_filename):
            mtime = os.path.getmtime(self.users_filename)
            if mtime > self.users_filename_last_loaded:
                logger.info("Detected update of users file, reloading.")
                self.load_users(self.users_filename)

    def authenticate(self, username, password):
        if not hasher:
            raise NoHasherAvailableError()
        self.reload_users()
        if username not in self.users:
            # User does not exist.
            return None
        else:
            if hasher.check_password(password, self.users[username]):
                return True
            return False

def require(function):
    """ Decorator to require authentication. """

    def wrapper(self, *args, **kwargs):

        if not authenticator:
            return function(self, *args, **kwargs)

        if not hasher:
            self.set_header("content-type", "text/plain")
            self.set_status(500)
            self.write("error: no password hasher available: install py-bcrypt")
            return

        # First look for our cookie.
        session_cookie = self.get_secure_cookie("dumpy-session")
        if session_cookie:
            self.session["username"] = session_cookie
            logger.debug("Found existing session for %s" % (
                    self.session["username"]))
            return function(self, *args, **kwargs)

        # Do we have an authentication header?
        auth_header = self.get_authorization()
        if auth_header:
            username, password = auth_header
            ok = authenticator.authenticate(username, password)
            if ok is False:
                logger.warn("Failed to login user %s: bad password" % (
                        username))
                self.write("bad password")
            elif ok is None:
                logger.warn("Failed to login user %s: unknown username" % (
                        username))
                self.write("unknown user")
            elif ok is True:
                logger.warn("User %s logged in." % (username))
                self.set_secure_cookie("dumpy-session", username)
                return function(self, *args, **kwargs)

        # Time to error out.
        self.set_status(401)
        self.set_header('WWW-Authenticate', 'Basic realm=Restricted')
        self.finish()
        return

    return wrapper

if has_bcrypt:
    hasher = BcryptHasher()
else:
    hasher = None

authenticator = Authenticator()
