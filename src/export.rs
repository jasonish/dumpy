// SPDX-FileCopyrightText: (C) 2022 Jason Ish <jason@codemonkey.net>
// SPDX-License-Identifier: MIT

use anyhow::Result;
use clap::Parser as ClapParser;
use lazy_static::lazy_static;
use regex::Regex;
use std::collections::HashMap;
use std::ffi::OsStr;
use std::path::{Path, PathBuf};
use std::time::UNIX_EPOCH;
use tracing::{debug, error, Level};
use tracing::{info, warn};

type SortedFiles = HashMap<(String, u64), Vec<(u64, PathBuf)>>;

const MIN_ROTATION_TIMESTAMP: u64 = 946_684_800; // 2000-01-01 00:00:00 UTC
const MAX_ROTATION_TIMESTAMP: u64 = 4_102_444_800; // 2100-01-01 00:00:00 UTC

#[derive(Debug, ClapParser)]
pub(crate) struct ExportArgs {
    /// Directory to process
    #[clap(long)]
    directory: String,

    /// Optional PCAP filter
    #[clap(long)]
    filter: Option<String>,

    /// Filename prefix
    ///
    /// If set, only files beginning with the provided prefix will be processed.
    #[clap(long)]
    prefix: Option<String>,

    /// Recurse through --directory
    #[clap(long)]
    recursive: bool,

    /// Start timestamp in seconds
    ///
    /// If provided, packets with a timestamp older than the one provided will
    /// not be exported.
    #[clap(long)]
    start_time: Option<i64>,

    /// Duration after start-time
    ///
    /// Sets the duration in seconds after start-time that packets will be
    /// considered for export. This refers to the packet timestamp, not real
    /// time.
    #[clap(long)]
    duration: Option<i64>,

    /// Enable more verbose logging
    #[clap(long, short, action = clap::ArgAction::Count)]
    verbose: u8,

    /// Output filename
    ///
    /// Filename to output packets to. "-" can be used for stdout.
    #[clap(long)]
    output: String,

    /// Output logs in JSON.
    #[clap(long)]
    json: bool,
}

pub(crate) fn main(args: ExportArgs) -> anyhow::Result<()> {
    let level = if args.verbose == 0 {
        Level::ERROR
    } else if args.verbose == 1 {
        Level::INFO
    } else {
        Level::DEBUG
    };

    crate::logging::init_logging(level, args.json);

    info!("args: {:?}", std::env::args());

    let mut dump_out = None;

    if let Ok(Some(files)) = load_files(Path::new(&args.directory), &args) {
        for (_id, files) in files {
            for (_ts, filename) in files {
                if let Err(err) = process_file(&args, &filename, &mut dump_out) {
                    error!("Failed to process {} -- {:?}", filename.display(), err);
                    std::process::exit(1);
                }
            }
        }
    } else {
        warn!("Failed to loaded sorted file list, processing will not be optimized");
        process_dir(&args, Path::new(&args.directory), &mut dump_out);
    }
    Ok(())
}

fn load_files(directory: &Path, args: &ExportArgs) -> Result<Option<SortedFiles>> {
    // Recursive not yet supported.
    if args.recursive {
        info!("Optimized file sorting not supported for recursive mode yet");
        return Ok(None);
    }
    let start_time = args.start_time.unwrap_or(0) as u64;
    let end_time = start_time + args.duration.unwrap_or(i64::MAX) as u64;

    let mut sorted: SortedFiles = HashMap::new();
    for entry in std::fs::read_dir(directory)? {
        let entry = entry?;
        let path = entry.path();
        if path.is_file() {
            if let Some(prefix) = &args.prefix {
                if let Some(filename) = path.file_name().and_then(|filename| filename.to_str()) {
                    if !filename.starts_with(prefix) {
                        continue;
                    }
                }
            }
            let filename = path.file_name().unwrap();
            if let Some((basename, id, ts)) = parse_filename(filename) {
                let list = sorted.entry((basename, id)).or_default();
                list.push((ts, path));
            } else {
                // Get out of here, we can't present a fully sorted file set.
                return Ok(None);
            }
        }
    }

    for files in sorted.values_mut() {
        files.sort_by_key(|a| a.0);

        let tmp = files.clone();
        let mut pit = tmp.iter().peekable();
        files.retain(|e| {
            let _current = pit.next().unwrap();
            if e.0 > end_time {
                debug!("Removing {}, it starts after the end time", e.1.display());
                return false;
            }
            if let Some(next) = pit.peek() {
                if next.0 < start_time {
                    // The next file in our sorted list has a creating time less than our start
                    // time, we can remove this file.
                    debug!("Removing {}, it ends before our start time", e.1.display());
                    return false;
                }
            }
            true
        });
    }

    Ok(Some(sorted))
}

fn parse_filename(filename: &OsStr) -> Option<(String, u64, u64)> {
    lazy_static! {
        // Matches the Suricata pcap-log format with thread-id.
        static ref RE_SURICATA_WITH_THREAD_ID: Regex =
            Regex::new(r"^(.*?)\.(\d+)(\.(\d+))?$").unwrap();

        // Matches a simpler format where the filename just contains a
        // timestamp.
        static ref RE_SIMPLE: Regex = Regex::new(r"(\d+)").unwrap();
    }
    if let Some(filename) = filename.to_str() {
        if let Some(m) = RE_SURICATA_WITH_THREAD_ID.captures(filename) {
            let basename = m.get(1)?.as_str().to_owned();
            let a = m.get(2)?.as_str().parse::<u64>().ok()?;
            let b = if let Some(m) = m.get(4) {
                m.as_str().parse::<u64>().ok()?
            } else {
                0
            };
            let (id, ts) = if a > b { (b, a) } else { (a, b) };
            if (MIN_ROTATION_TIMESTAMP..=MAX_ROTATION_TIMESTAMP).contains(&ts) {
                return Some((basename, id, ts));
            }
            return None;
        }

        if let Some(m) = RE_SIMPLE.captures(filename) {
            let timestamp_match = m.get(1)?;
            let ts = timestamp_match.as_str().parse::<u64>().ok()?;
            if (MIN_ROTATION_TIMESTAMP..=MAX_ROTATION_TIMESTAMP).contains(&ts) {
                let basename = format!(
                    "{}{}",
                    &filename[..timestamp_match.start()],
                    &filename[timestamp_match.end()..]
                );
                return Some((basename, 0, ts));
            }
        }
    }

    None
}

fn process_dir(args: &ExportArgs, directory: &Path, out: &mut Option<pcap::Savefile>) {
    for entry in std::fs::read_dir(directory).unwrap() {
        let entry = entry.unwrap();
        let path = entry.path();
        if path.is_dir() && args.recursive {
            process_dir(args, &path, out);
        } else if path.is_file() {
            if let Some(prefix) = &args.prefix {
                if let Some(filename) = path.file_name().and_then(|filename| filename.to_str()) {
                    if !filename.starts_with(prefix) {
                        continue;
                    }
                }
            }
            if let Err(err) = process_file(args, &path, out) {
                error!("{:?}", err);
                std::process::exit(1);
            }
        } else {
            debug!("Ignoring {:?}", &path);
        }
    }
}

fn process_file(args: &ExportArgs, path: &Path, out: &mut Option<pcap::Savefile>) -> Result<()> {
    let metadata = std::fs::metadata(path)?;
    let mtime = metadata.modified()?.duration_since(UNIX_EPOCH)?.as_secs();
    if let Some(start_time) = args.start_time {
        if mtime < start_time as u64 {
            debug!(
                "Ignoring {}, last modified before {}",
                &path.display(),
                start_time
            );
            return Ok(());
        }
    }
    info!("Processing file {}", &path.display());
    let mut cf = pcap::Capture::from_file(path).unwrap();
    if let Some(filter) = &args.filter {
        cf.filter(filter, true)?;
    }
    loop {
        let n = cf.next_packet();
        match n {
            Ok(pkt) => {
                let secs = pkt.header.ts.tv_sec;
                if let Some(start_time) = args.start_time {
                    if secs < start_time {
                        continue;
                    }
                    if let Some(duration) = args.duration {
                        if secs > start_time + duration {
                            return Ok(());
                        }
                    }
                }
                if out.is_none() {
                    // A bit of a hack, but prevents writing any data until we know we've found
                    // a packet.
                    let cf0 = pcap::Capture::from_file(path).unwrap();
                    *out = Some(cf0.savefile(&args.output).unwrap());
                }
                out.as_mut().unwrap().write(&pkt);
            }
            Err(err) => {
                match err {
                    pcap::Error::NoMorePackets => {
                        break;
                    }
                    pcap::Error::PcapError(ref error)
                        // Truncation errors are expected.
                        if error.contains("truncated") => {
                            break;
                        }
                    _ => {}
                }
                warn!("pcap-error: {}: {}", path.display(), err);
                break;
            }
        }
    }
    Ok(())
}

#[cfg(test)]
mod test {
    use super::*;
    use std::ffi::OsString;

    #[test]
    fn test_parse_filename_with_thread() {
        let filename: OsString = "log.pcap.1.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("log.pcap".to_string(), 1, 1649915733));

        let filename: OsString = "log.pcap.11.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("log.pcap".to_string(), 11, 1649915733));

        let filename: OsString = "log.pcap.111.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("log.pcap".to_string(), 111, 1649915733));
    }

    #[test]
    fn test_parse_filename_without_thread() {
        let filename: OsString = "log.pcap.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("log.pcap".to_string(), 0, 1649915733));
    }

    #[test]
    fn test_parse_filename_simple_timestamp() {
        let filename: OsString = "trace-1649915733.pcap".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("trace-.pcap".to_string(), 0, 1649915733));
    }

    #[test]
    fn test_parse_filename_rejects_date_style_rotation() {
        assert!(parse_filename(&OsString::from("trace-20260701.pcap")).is_none());
        assert!(parse_filename(&OsString::from("trace-20260701120000.pcap")).is_none());
    }

    #[test]
    fn test_parse_filename_uses_trailing_rotation_fields() {
        let filename: OsString = "10.0.0.1.pcap.1.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, ("10.0.0.1.pcap".to_string(), 1, 1649915733));

        assert!(parse_filename(&OsString::from("trace.2026.07.03.pcap")).is_none());
    }

    #[test]
    fn test_load_files_groups_by_basename_and_thread() {
        let directory = std::env::temp_dir().join(format!(
            "dumpy-export-test-{}-{}",
            std::process::id(),
            std::thread::current().name().unwrap_or("unnamed")
        ));
        std::fs::create_dir_all(&directory).unwrap();

        for filename in [
            "log.pcap.1.1700001000",
            "log.pcap.1.1700002000",
            "log.pcap.1.1700003000",
            "report.pcap.1.1700001200",
        ] {
            std::fs::File::create(directory.join(filename)).unwrap();
        }

        let args = ExportArgs {
            directory: directory.to_string_lossy().into_owned(),
            filter: None,
            prefix: None,
            recursive: false,
            start_time: Some(1_700_001_250),
            duration: Some(250),
            verbose: 0,
            output: "unused.pcap".to_string(),
            json: false,
        };
        let files = load_files(&directory, &args).unwrap().unwrap();

        let log_files = &files[&("log.pcap".to_string(), 1)];
        assert_eq!(log_files.len(), 1);
        assert_eq!(
            log_files[0].1.file_name().unwrap(),
            OsStr::new("log.pcap.1.1700001000")
        );
        assert_eq!(files[&("report.pcap".to_string(), 1)].len(), 1);

        std::fs::remove_dir_all(directory).unwrap();
    }
}
