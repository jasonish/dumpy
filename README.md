# Dumpy

Dumpy is a simple to install, and simple to use web front end for PCAP
spool file directories such as those produced by daemonlogger or
tcpdump.

## Setup

1. First, setup at least one pcap spool directory.  This can be done
  using daemonlogger
  (http://www.snort.org/snort-downloads/additional-downloads#daemonlogger)
  or other tools like it.  Example command:

      daemonlogger -i eth0 -l /data/capture -s 1000000000 -M 70 -r

2. Download a dumpy binary package
  (https://github.com/jasonish/dumpy/releases) or build from source.
  Note: Requires libpcap to be installed.

3. Configure:

	* Add a spool.  The following command will create a spool named
      "default" (note: the name default has no special meaning), with
      directory /data/capture and a filename prefix of
      daemonlogger.pcap - this matches the use of daemonlogger above):

			./dumpy config spool add -name default -directory /data/capture \
			    -prefix daemonlogger.pcap

    * Add a user:

			./dumpy config passwd username password

4. Start dumpy:

		./dumpy start

5. Then point your browser at http://<hostname>:7000/

## Building

Building Dumpy requires:
- A working Go 1.9+ development environment.
- A working [Node.js](https://nodejs.org/en/) install with npm for frontend
  dependencies.

Then:

1. Install dependencies:

    make install-deps

2. Build

    make
