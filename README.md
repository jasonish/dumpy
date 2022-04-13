# Dumpy

Dumpy is a simple to install, and simple to use web frontend for PCAP spool 
file directories such as those produced by Suricata.

## Setup

1. First configure and start a tool like Suricata, or daemonlogger to write 
   PCAP files to a directory such as `/data/capture`.
2. Download a Dumpy binary package (https://github.com/jasonish/dumpy/releases) 
   or build from source. Note: Requires libpcap to be installed.
3. Configure:
   1. Choose a directory where the `dumpy.yaml` configuration file will 
      exist and change to that directory. For now, lets use `~/dumpy`.
   2. Tell Dumpy where to find the PCAP directory using the `dumpy config` 
      command, for example:
      ```
      dumpy config spool add default /data/capture
      ```
   3. Optionally add a user, if you don't authentication won't be required.
      ```
      dumpy config passwd username password
      ```
4. Start Dumpy:
   ```
   dumpy server
   ```
5. Then point your browser at http://<hostname>:7000/

## Building

Building Dumpy requires Rust and Cargo to be install, then simply:
```
cargo build
```

## TLS

TLS can be enabled through the `dumpy config` command but you will first 
need TLS certificate and key files.

A self-signed certificate and key and be created with openssl:

```
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 
      -days 365 -nodes -subj '/CN=localhost'
```

Then TLS can be enabled in Dumpy with the following command:
```
dumpy config set tls.cert cert.pem
dumpy config set tls.key cert.key
dumpy config set tls.enabled true
```
