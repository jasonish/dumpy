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

import sys
import os
import unittest
import glob

sys.path.insert(0, os.path.abspath(os.path.dirname(__file__) + "/../.."))

test_modules = []
for module in glob.glob("%s/test_*.py" % (os.path.dirname(__file__))):
    test_modules.append("dumpy.tests.%s" % (
            os.path.basename(module).rsplit(".", 1)[0]))

def main(args=sys.argv[1:]):
    if args:
        tests = []
        for test in test_modules:
            for arg in args:
                if test.find(arg) >= 0:
                    tests.append(test)
    else:
        tests = test_modules

    suite = unittest.defaultTestLoader.loadTestsFromNames(tests)
    unittest.TextTestRunner(verbosity=2).run(suite)

if __name__ == "__main__":
    sys.exit(main())
    