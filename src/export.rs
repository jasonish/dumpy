// SPDX-License-Identifier: MIT
//
// Copyright (C) 2022 Jason Ish

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

type SortedFiles = HashMap<u64, Vec<(u64, PathBuf)>>;

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
            if let Some((id, ts)) = parse_filename(filename) {
                let list = sorted.entry(id).or_default();
                list.push((ts, path));
            } else {
                // Get out of here, we can't present a fully sorted file set.
                return Ok(None);
            }
        }
    }

    for files in sorted.values_mut() {
        files.sort_by(|a, b| a.0.cmp(&b.0));

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

fn parse_filename(filename: &OsStr) -> Option<(u64, u64)> {
    lazy_static! {
        static ref RE: Regex = Regex::new(r"(.*?)\.(\d+)(\.(\d+))?").unwrap();
    }
    if let Some(filename) = filename.to_str() {
        if let Some(m) = RE.captures(filename) {
            let a = m.get(2).unwrap().as_str().parse::<u64>().unwrap_or(0);
            let b = if let Some(m) = m.get(4) {
                m.as_str().parse::<u64>().unwrap_or(0)
            } else {
                0
            };
            let (id, ts) = if a > b { (b, a) } else { (a, b) };
            return Some((id, ts));
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
        let n = cf.next();
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
                    pcap::Error::PcapError(ref error) => {
                        // Truncation errors are expected.
                        if error.contains("truncated") {
                            break;
                        }
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
        assert_eq!(r, (1, 1649915733));

        let filename: OsString = "log.pcap.11.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, (11, 1649915733));

        let filename: OsString = "log.pcap.111.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, (111, 1649915733));
    }

    #[test]
    fn test_parse_filename_without_thread() {
        let filename: OsString = "log.pcap.1649915733".into();
        let r = parse_filename(&filename).unwrap();
        assert_eq!(r, (0, 1649915733));
    }
}
