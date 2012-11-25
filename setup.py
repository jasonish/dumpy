from distutils.core import setup
from distutils.command.install import install

from dumpy import version

# Override the install command as we are not currently installable.
# Maybe someday.
class deny_install(install):
    def run(self):
        print("error this package is not currently installable")

setup(
    name="dumpy",
    version=version.VERSION,
    packages=["dumpy"],
    scripts=["bin/dumpy-web", 
             "bin/dumpy-extract",
             "bin/dumpy-passwd"],
    cmdclass={"install": deny_install},
    )

