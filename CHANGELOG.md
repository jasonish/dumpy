# CHANGELOG

## Unreleased
- Add command-line configuration options to `dumpy server` command
  - All arguments are now optional
  - Falls back to dumpy.yaml configuration when arguments not provided
  - Command-line arguments override configuration file values
  - Supports --directory, --name, --prefix, and --port options
- Improve PCAP filename parsing to support simpler timestamp-only formats
  - Previously only supported Suricata's thread-id.timestamp format
  - Now also supports files with just a timestamp in the filename

## 0.5.0 - 2025-01-06
- Add dark mode toggle and GitHub repository link to web interface
- Replace jQuery/Bootstrap with vanilla JS/CSS lightweight implementation
- Add purge command to clean up old PCAP files
- Update to AlmaLinux 9 in Dockerfile
- Update all dependencies to latest versions

## 0.4.2 - 2024-03-26
- Update dependencies
- Attempt to log in local time zone. If not possible, use UTC and suffix the 
  timestamp with "Z".
- Fix argument handling

## 0.4.1 - 2022-11-21

Update dependencies only.
