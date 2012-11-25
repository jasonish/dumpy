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
import io

try:
    import bcrypt
    has_bcrypt = True
except:
    has_bcrypt = False

from dumpy import auth

class HasherTests(unittest.TestCase):

    def test_bcrypt(self):
        if not has_bcrypt:
            sys.stderr.write("skipped--")
            return

        good_password = "password"
        bad_password = "p4ssword"
        hashed = bcrypt.hashpw(good_password, bcrypt.gensalt())
        self.assertEquals(hashed, bcrypt.hashpw(good_password, hashed))
        self.assertNotEqual(hashed, bcrypt.hashpw(bad_password, hashed))

    def test_bcrypt_hasher(self):
        if not has_bcrypt:
            sys.stderr.write("skipped--")
            return

        hasher = auth.BcryptHasher()
        good_password = "password"
        bad_password = "p4ssword"
        hashed = hasher.hash_password(good_password)
        self.assertTrue(hashed.startswith(hasher.prefix))
        self.assertTrue(hasher.check_password(good_password, hashed))
        self.assertFalse(hasher.check_password(bad_password, hashed))

class AuthenticatorTests(unittest.TestCase):

    password_file = u"""
# buser:password, bcrypt
user0:bcrypt:$2a$12$n3JSt.DTLyBj/KjHNZMy2up8r2k.6z0y6fd49WNr3hJHX9q1xoiI2
"""

    def test_load_users(self):
        authenticator = auth.Authenticator()
        authenticator.load_users(None, io.StringIO(self.password_file))
        self.assertTrue("user0" in authenticator.users)

    def test_bcrypt_auth(self):
        if not has_bcrypt:
            sys.stderr.write("skipped--")
            return
        authenticator = auth.Authenticator()
        authenticator.load_users(None, io.StringIO(self.password_file))
        self.assertFalse(authenticator.authenticate("user0", "bad-password"))
        self.assertTrue(authenticator.authenticate("user0", "password"))
