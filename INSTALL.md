# Dumpy - Installation

## Requirements

* Python 2.6 or 2.7
	- On Linux you most likely already have this installed.
* Tornado Web Server
	- CentOS w/EPEL: python-tornado
	- Fedora: python-tornado
	- Ubuntu: python-tornado
* py-bcrypt (for authentication)
	- CentOS w/EPEL: py-bcrypt
	- Fedora: py-bcrypt
	- Ubuntu: python-bcrypt

Note that py-bcrypt is optional, but without Dumpy will not be able to
perform authentication.

## Initial Configuration

1) Initialize the configuration and other site specific files by
   running the command:

      make setup

2) Edit etc/config, in particular update the spool section to point to
   your pcap spool directory:

      [spool]
      directory = /data/capture
      prefix = daemonlogger.pcap

### Users and Authentication

By default there is NO AUTHENTICATION.  Authentication can be enabled
by creating a users file and adding a user.  This can be done with the
following command:

      ./bin/dumpy-passwd <username> <password> >> etc/users

### SSL (HTTPS)

HTTPS can be enabled by providing SSL key and certificate files.  The
relevant portion of etc/config is:

      [http]
      ssl = no
      ssl-key = etc/ssl/server.key
      ssl-cert = etc/ssl/server.crt

If you just want to use a self signed certificate then run `make ssl`
and then set the `ssl` option to `yes`.

## Running

The Dumpy server can be started by running:

      ./bin/dumpy-web

