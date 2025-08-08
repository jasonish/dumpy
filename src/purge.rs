// SPDX-FileCopyrightText: (C) 2025 Jason Ish <jason@codemonkey.net>
// SPDX-License-Identifier: MIT

use anyhow::{bail, Result};
use clap::Parser;
use std::path::{Path, PathBuf};
use std::time::{Duration, SystemTime};
use tokio::time::sleep;
use tracing::{error, info, warn};

#[derive(Debug, Parser)]
pub struct PurgeArgs {
    /// Directory containing PCAP files to purge
    directory: String,

    /// Keep only the newest N files
    #[clap(long, conflicts_with = "max_size")]
    keep_files: Option<usize>,

    /// Keep only the newest files up to this total size (e.g., "10G", "500M")
    #[clap(long, conflicts_with = "keep_files")]
    max_size: Option<String>,

    /// Only process files with this prefix
    #[clap(long)]
    prefix: Option<String>,

    /// Actually delete files (default is dry-run)
    #[clap(long, short)]
    force: bool,

    /// Run purge every N minutes (for container deployments)
    #[clap(long, short)]
    interval: Option<u64>,
}

#[derive(Debug)]
struct FileInfo {
    path: PathBuf,
    size: u64,
    modified: SystemTime,
}

pub fn purge_main(args: PurgeArgs) -> Result<()> {
    tracing_subscriber::fmt()
        .with_writer(std::io::stderr)
        .init();

    if let Some(interval_minutes) = args.interval {
        info!("Starting purge with {}-minute interval", interval_minutes);
        let rt = tokio::runtime::Runtime::new()?;
        rt.block_on(async {
            loop {
                if let Err(e) = run_purge_once(&args).await {
                    error!("Purge failed: {}", e);
                }

                info!("Sleeping for {} minutes", interval_minutes);
                sleep(Duration::from_secs(interval_minutes * 60)).await;
            }
        })
    } else {
        let rt = tokio::runtime::Runtime::new()?;
        rt.block_on(run_purge_once(&args))
    }
}

async fn run_purge_once(args: &PurgeArgs) -> Result<()> {
    let files = collect_pcap_files(&args.directory, args.prefix.as_deref())?;
    if files.is_empty() {
        info!("No PCAP files found in directory '{}'", args.directory);
        return Ok(());
    }

    let files_to_delete = if let Some(keep_count) = args.keep_files {
        calculate_files_to_delete_by_count(&files, keep_count)
    } else if let Some(max_size_str) = &args.max_size {
        let max_size_bytes = parse_size(max_size_str)?;
        calculate_files_to_delete_by_size(&files, max_size_bytes)
    } else {
        bail!("Either --keep-files or --max-size must be specified");
    };

    if files_to_delete.is_empty() {
        info!("No files need to be deleted");
        return Ok(());
    }

    let total_size: u64 = files_to_delete.iter().map(|f| f.size).sum();
    let total_size_mb = total_size as f64 / 1024.0 / 1024.0;

    if !args.force {
        info!(
            "Would delete {} files ({:.2} MB)",
            files_to_delete.len(),
            total_size_mb
        );
        for file in &files_to_delete {
            info!("  {}", file.path.display());
        }
        warn!("To actually delete these files, run with --force");
    } else {
        info!(
            "Deleting {} files ({:.2} MB)",
            files_to_delete.len(),
            total_size_mb
        );

        let mut delete_count = 0;
        let mut delete_errors = 0;

        for file in &files_to_delete {
            match std::fs::remove_file(&file.path) {
                Ok(_) => {
                    delete_count += 1;
                    info!("Deleted: {}", file.path.display());
                }
                Err(e) => {
                    delete_errors += 1;
                    error!("Failed to delete {}: {}", file.path.display(), e);
                }
            }
        }

        info!(
            "Deleted {} files successfully, {} errors",
            delete_count, delete_errors
        );
    }

    Ok(())
}

fn collect_pcap_files(directory: &str, prefix: Option<&str>) -> Result<Vec<FileInfo>> {
    let mut files = Vec::new();
    let dir_path = Path::new(directory);

    if !dir_path.exists() {
        bail!("Directory does not exist: {}", directory);
    }

    if !dir_path.is_dir() {
        bail!("Path is not a directory: {}", directory);
    }

    for entry in std::fs::read_dir(dir_path)? {
        let entry = entry?;
        let path = entry.path();

        // Check if it's a regular file without following symlinks
        if !entry.file_type()?.is_file() {
            continue;
        }

        // Skip non-PCAP files
        if let Some(filename) = path.file_name().and_then(|f| f.to_str()) {
            // Check for proper PCAP file extensions
            let is_pcap = filename.ends_with(".pcap")
                || filename.ends_with(".pcapng")
                || filename.ends_with(".cap");

            if !is_pcap {
                continue;
            }

            if let Some(prefix) = prefix {
                if !filename.starts_with(prefix) {
                    continue;
                }
            }
        }

        let metadata = entry.metadata()?;
        files.push(FileInfo {
            path,
            size: metadata.len(),
            modified: metadata.modified()?,
        });
    }

    // Sort by modification time, newest first
    files.sort_by(|a, b| b.modified.cmp(&a.modified));

    Ok(files)
}

fn calculate_files_to_delete_by_count(files: &[FileInfo], keep_count: usize) -> Vec<&FileInfo> {
    if files.len() <= keep_count {
        return Vec::new();
    }

    files[keep_count..].iter().collect()
}

fn calculate_files_to_delete_by_size(files: &[FileInfo], max_size_bytes: u64) -> Vec<&FileInfo> {
    let mut total_size = 0u64;
    let mut keep_index = 0;

    for (idx, file) in files.iter().enumerate() {
        if total_size + file.size > max_size_bytes {
            break;
        }
        total_size += file.size;
        keep_index = idx + 1;
    }

    files[keep_index..].iter().collect()
}

fn parse_size(size_str: &str) -> Result<u64> {
    let size_str = size_str.trim().to_uppercase();

    if let Some(num_str) = size_str.strip_suffix('G') {
        let num: f64 = num_str.parse()?;
        Ok((num * 1024.0 * 1024.0 * 1024.0) as u64)
    } else if let Some(num_str) = size_str.strip_suffix('M') {
        let num: f64 = num_str.parse()?;
        Ok((num * 1024.0 * 1024.0) as u64)
    } else if let Some(num_str) = size_str.strip_suffix('K') {
        let num: f64 = num_str.parse()?;
        Ok((num * 1024.0) as u64)
    } else {
        let num: u64 = size_str.parse()?;
        Ok(num)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_size() {
        assert_eq!(parse_size("1024").unwrap(), 1024);
        assert_eq!(parse_size("1K").unwrap(), 1024);
        assert_eq!(parse_size("1M").unwrap(), 1024 * 1024);
        assert_eq!(parse_size("1G").unwrap(), 1024 * 1024 * 1024);
        assert_eq!(
            parse_size("1.5G").unwrap(),
            (1.5 * 1024.0 * 1024.0 * 1024.0) as u64
        );
    }
}
