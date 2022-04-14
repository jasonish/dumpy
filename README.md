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

## Other Installation Options

### With Cargo

```
cargo install --git https://github.com/jasonish/dumpy
```

Installation from *crates.io* will become available after 0.4.0 is released.

## Suricata Configuration

For Dumpy to be of much use you will need a tool to log PCAP files. Suricata 
can be configured to do this with the `pcap-log` output:

```yaml
  - pcap-log:
      enabled: yes
      filename: log.pcap
      limit: 256mb
      max-files: 1000
      compression: none
      mode: normal
      dir: /data/capture
```

Or using multi-threaded mode where each worker thread will write to its own 
file in hopes to improve performance:

```yaml
  - pcap-log:
      enabled: yes
      filename: log.pcap.%n.%t
      limit: 256mb
      max-files: 250
      compression: none
      mode: multi
      dir: /data/capture
```

Optimizations exist for processing directories with the filename patterns 
above, however most any patterns should work, however Dumpy may not be able 
to eliminate files from being read if the above patterns are not followed.

## Alternative: tcpdump

Even `tcpdump` can be used to generate a *spool* directory of PCAP files:

```
tcpdump -w /data/captures/pcap.%s -G 3600 -s0 -i enp10s0
```

Note the `-G` parameter and the `%s` in the filename. With the above command 
`tcpdump` will open a new files every hour and the filename will be prefixed 
with the unix timestamp in seconds.

Note: You will have to take care of cleaning up old files.

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
