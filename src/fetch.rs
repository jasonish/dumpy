// SPDX-License-Identifier: MIT
//
// Copyright (C) 2022 Jason Ish

use crate::config::{Config, SpoolConfig};
use anyhow::{anyhow, bail};
use axum::body::StreamBody;
use axum::extract::{Form, Query};
use axum::http::{HeaderMap, HeaderValue, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::Extension;
use serde::Deserialize;
use std::ops::Add;
use std::ops::Sub;
use std::process::Stdio;
use thiserror::Error;
use time::format_description::well_known::Rfc3339;
use time::{Duration, OffsetDateTime};
use tokio::io::AsyncBufReadExt;
use tokio::io::AsyncReadExt;
use tokio::io::BufReader;
use tokio_stream::wrappers::UnboundedReceiverStream;
use tracing::{error, info};

const DEFAULT_DURATION: &str = "1m";

#[derive(Deserialize, Debug, Clone)]
#[allow(dead_code)]
pub struct FetchRequest {
    #[serde(rename = "query-type")]
    query_type: String,
    filter: Option<String>,
    #[serde(rename = "start-time")]
    start_time: Option<String>,
    duration: Option<String>,
    event: Option<String>,
    #[serde(rename = "duration-before")]
    duration_before: Option<String>,
    #[serde(rename = "duration-after")]
    duration_after: Option<String>,
    spool: String,
}

#[derive(Debug, Error)]
#[error("{reason:?}")]
pub struct BadRequest {
    reason: String,
}

impl BadRequest {
    fn new<S: AsRef<str>>(reason: S) -> Self {
        Self {
            reason: reason.as_ref().to_string(),
        }
    }
}

impl IntoResponse for BadRequest {
    fn into_response(self) -> Response {
        (StatusCode::BAD_REQUEST, self.reason).into_response()
    }
}

fn get_spool_config<'a>(config: &'a Config, name: &str) -> Option<&'a SpoolConfig> {
    for spool in &config.spools {
        if spool.name == name {
            return Some(spool);
        }
    }
    None
}

pub async fn fetch_get(
    Extension(config): Extension<Config>,
    Query(params): Query<FetchRequest>,
) -> Result<impl IntoResponse, BadRequest> {
    fetch(config, params).await
}

pub async fn fetch_post(
    Extension(config): Extension<Config>,
    Form(params): Form<FetchRequest>,
) -> Result<impl IntoResponse, BadRequest> {
    fetch(config, params).await
}

pub async fn fetch(config: Config, params: FetchRequest) -> Result<impl IntoResponse, BadRequest> {
    let spool_config = get_spool_config(&config, &params.spool).ok_or_else(|| {
        error!("No spool with the name: {}", &params.spool);
        BadRequest::new(&format!("invalid spool name: {}", &params.spool))
    })?;

    let start_time;
    let duration;
    let filter;
    let filename;

    match params.query_type.as_ref() {
        "filter" => {
            let stime = match params.start_time.as_ref() {
                Some(stime) => stime,
                None => {
                    return Err(BadRequest::new("start-time is required"));
                }
            };

            start_time = OffsetDateTime::parse(stime, &Rfc3339).map_err(|err| {
                error!("Invalid start-time: {} -- {:?}", &stime, err);
                BadRequest::new("invalid start-time")
            })?;
            let pduration = params.duration.as_deref().unwrap_or(DEFAULT_DURATION);
            duration = parse_duration(pduration).map_err(|err| {
                error!("Invalid duration: {} -- {:?}", &pduration, err);
                BadRequest::new("invalid duration")
            })?;
            filter = params.clone().filter.unwrap();
            filename = format!("{}.pcap", start_time.unix_timestamp());
        }
        "event" => {
            if let Some(event) = &params.event {
                let event: EveEvent = serde_json::from_str(event).map_err(|err| {
                    error!("Failed to decode event: {:?} -- event={}", err, &event);
                    BadRequest::new("bad event")
                })?;

                (start_time, duration) = parse_event_timeframe(&event, &params)?;

                if event.src_port.is_some() && event.dest_port.is_some() {
                    filter = format!(
                        "{} and ((host {} and port {}) and (host {} and port {}))",
                        &event.proto.to_lowercase(),
                        &event.src_ip,
                        event.src_port.as_ref().unwrap(),
                        &event.dest_ip,
                        event.dest_port.as_ref().unwrap(),
                    );
                } else {
                    filter = format!(
                        "{} and host {} and host {}",
                        &event.proto, &event.src_ip, &event.dest_ip
                    );
                }

                // Create a filename based on event details.
                let signature_id = event.alert.map(|a| a.signature_id).unwrap_or(0);
                filename = format!(
                    "{}-{}-{}-{}-{}-{}.pcap",
                    signature_id,
                    start_time.unix_timestamp(),
                    &event.src_ip,
                    event.src_port.unwrap_or(0),
                    &event.dest_ip,
                    event.dest_port.unwrap_or(0)
                );
            } else {
                return Err(BadRequest::new("no event provided for event query"));
            }
        }
        _ => {
            error!("Invalid query type: {}", &params.query_type);
            return Err(BadRequest::new("bad query type"));
        }
    }

    let process_name = std::env::args().next().as_ref().unwrap().to_string();
    let mut command = tokio::process::Command::new(process_name);
    command
        .arg("export")
        .arg("--json")
        .arg("-v")
        .arg("--directory")
        .arg(&spool_config.directory)
        .arg("--filter")
        .arg(filter)
        .arg("--start-time")
        .arg(format!("{}", start_time.unix_timestamp()))
        .arg("--duration")
        .arg(format!("{}", duration.as_seconds_f64() as i64))
        .arg("--output")
        .arg("-");
    if let Some(prefix) = &spool_config.prefix {
        command.arg("--prefix").arg(prefix);
    }
    command.stdout(Stdio::piped());
    command.stderr(Stdio::piped());
    let mut child = command.spawn().unwrap();
    let mut stdout = child.stdout.take().unwrap();
    let stderr = child.stderr.take().unwrap();

    let (tx, rx) =
        tokio::sync::mpsc::unbounded_channel::<std::result::Result<Vec<u8>, std::io::Error>>();
    let rx = UnboundedReceiverStream::new(rx);
    let body = StreamBody::new(rx);
    let mut stderr_reader = BufReader::new(stderr).lines();
    let (wait_tx, wait_rx) = tokio::sync::oneshot::channel::<&str>();

    tokio::spawn(async move {
        let mut bytes = 0;
        let mut wait_tx = Some(wait_tx);
        let mut client_closed = false;
        loop {
            let mut buf = Vec::with_capacity(8192);
            tokio::select! {
                _ = child.wait() => {
                    break;
                }
                _ = tx.closed() => {
                    client_closed = true;
                    break;
                }
                closed = read_stderr(&mut stderr_reader) => {
                    if closed {
                        break;
                    }
                }
                x = stdout.read_buf(&mut buf) => {
                    match x {
                        Ok(n) => {
                            if n == 0 {
                                break;
                            } else {
                                if let Some(wait_tx) = wait_tx.take() {
                                    wait_tx.send("ok").unwrap();
                                }
                                if tx.send(Ok(buf)).is_err() {
                                    error!("Failed to write to body stream, client must have closed connection");
                                    client_closed = true;
                                    break;
                                }
                                bytes += n;
                            }
                        }
                        Err(err) => {
                            error!("Error reading export process stdout: {:?}", err);
                            break;
                        }
                    }
                }
            }
        }

        if client_closed {
            let _ = child.start_kill();
        }
        let status = child.wait().await.unwrap();
        if status.success() {
            info!(
                "Export process exited successfully, bytes written: {}",
                bytes
            );
        } else {
            error!("Export process exited with error code: {:?}", status);
        }
        if let Some(wait_tx) = wait_tx.take() {
            if status.success() {
                wait_tx.send("nopkt").unwrap();
            } else {
                wait_tx.send("err").unwrap();
            }
        }
    });

    Ok(match wait_rx.await {
        Err(err) => {
            error!("Error on wait channel: {:?}", err);
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                "An error occurred. See server logs for details",
            )
                .into_response()
        }
        Ok(status) => {
            if status == "ok" {
                let mut headers = HeaderMap::new();
                headers.insert(
                    axum::http::header::CONTENT_TYPE,
                    HeaderValue::from_str("application/vnd.tcpdump.pcap").unwrap(),
                );
                headers.insert(
                    axum::http::header::CONTENT_DISPOSITION,
                    HeaderValue::from_str(&format!("attachment; filename={}", filename)).unwrap(),
                );

                (headers, body).into_response()
            } else if status == "nopkt" {
                (StatusCode::NOT_FOUND, "No packets found.").into_response()
            } else {
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "An error occurred. See server logs for details",
                )
                    .into_response()
            }
        }
    })
}

async fn read_stderr(
    reader: &mut tokio::io::Lines<tokio::io::BufReader<tokio::process::ChildStderr>>,
) -> bool {
    match reader.next_line().await {
        Ok(Some(next)) => {
            if let Ok(v) = serde_json::from_str::<serde_json::Value>(&next) {
                let level = v["level"].as_str().unwrap_or("INFO");
                if let Some(message) = v["fields"]["message"].as_str() {
                    match level {
                        "ERROR" => error!("{}", message),
                        _ => info!("{}", message),
                    }
                    return false;
                }
            }
            error!("{}", &next);
            false
        }
        Ok(None) | Err(_) => true,
    }
}

fn parse_duration(s: &str) -> anyhow::Result<Duration> {
    let re = regex::Regex::new(r"^(\d+)(.+)$")?;
    let captures = re
        .captures(s)
        .ok_or_else(|| anyhow!("invalid duration string: {}", s))?;
    let value = captures.get(1).unwrap().as_str().parse::<i64>()?;
    let unit = captures.get(2).unwrap().as_str();
    match unit {
        "s" => Ok(time::Duration::seconds(value)),
        "m" => Ok(time::Duration::minutes(value)),
        "h" => Ok(time::Duration::hours(value)),
        _ => bail!("invalid duration unit: {}", unit),
    }
}

fn parse_timestamp(s: &str) -> Result<OffsetDateTime, time::Error> {
    let re = regex::Regex::new(r"^(?P<leader>.*?)(?P<sep>[+-])(?P<H>\d\d)(?P<M>\d\d)").unwrap();
    let fixed = re.replace_all(s, r"$leader$sep$H:$M");
    let parsed = OffsetDateTime::parse(&fixed, &Rfc3339)?;
    Ok(parsed)
}

/// Parse out a start time and a duration from the flow or netflow fields.
///
/// The times are adjusted to be one second before and one second after the provided
/// start and end times which seems to be required to catch the complete flow.
fn parse_flow_timeframe(flow: &EveEventFlow) -> Option<(OffsetDateTime, Duration)> {
    if let Ok(start_time) = parse_timestamp(&flow.start) {
        let adjusted_start_time = start_time - Duration::seconds(1);
        if let Some(end_time) = &flow.end {
            if let Ok(end_time) = parse_timestamp(end_time) {
                let duration = (end_time - start_time) + Duration::seconds(2);
                return Some((adjusted_start_time, duration));
            }
        }
        return Some((
            adjusted_start_time,
            parse_duration(DEFAULT_DURATION).unwrap(),
        ));
    }
    None
}

fn parse_event_timeframe(
    event: &EveEvent,
    params: &FetchRequest,
) -> Result<(OffsetDateTime, Duration), BadRequest> {
    if event.event_type == "flow" {
        if let Some(flow) = &event.flow {
            if let Some(timeinfo) = parse_flow_timeframe(flow) {
                return Ok(timeinfo);
            }
        }
    } else if event.event_type == "netflow" {
        if let Some(netflow) = &event.netflow {
            if let Some(timeinfo) = parse_flow_timeframe(netflow) {
                return Ok(timeinfo);
            }
        }
    }

    let timestamp = parse_timestamp(&event.timestamp).map_err(|err| {
        error!(
            "Failed to parse event timestamp: {} -- {:?}",
            &event.timestamp, err
        );
        BadRequest::new("failed to parse event timestamp")
    })?;

    let duration_before = params
        .duration_before
        .as_deref()
        .unwrap_or(DEFAULT_DURATION);
    let duration_before =
        parse_duration(duration_before).map_err(|_| BadRequest::new("invalid duration-before"))?;

    let duration_after = params.duration_after.as_deref().unwrap_or(DEFAULT_DURATION);
    let duration_after =
        parse_duration(duration_after).map_err(|_| BadRequest::new("invalid duration-after"))?;

    let start_time = timestamp.sub(duration_before);
    let end_time = timestamp.add(duration_after);
    let duration = end_time.sub(start_time);

    Ok((start_time, duration))
}

#[derive(Debug, Deserialize)]
struct EveEvent {
    timestamp: String,
    event_type: String,
    proto: String,
    src_ip: String,
    src_port: Option<u16>,
    dest_ip: String,
    dest_port: Option<u16>,
    alert: Option<EveEventAlert>,
    flow: Option<EveEventFlow>,
    netflow: Option<EveEventFlow>,
}

#[derive(Debug, Deserialize)]
struct EveEventAlert {
    signature_id: u64,
}

#[derive(Debug, Deserialize)]
struct EveEventFlow {
    start: String,
    end: Option<String>,
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_parse_duration() {
        let duration = parse_duration("1s").unwrap();
        assert_eq!(duration, Duration::seconds(1));

        let duration = parse_duration("1m").unwrap();
        assert_eq!(duration, Duration::seconds(60));

        let duration = parse_duration("1h").unwrap();
        assert_eq!(duration, Duration::seconds(3600));

        assert!(parse_duration("1").is_err());
        assert!(parse_duration("1z").is_err());
        assert!(parse_duration("am").is_err());
        assert!(parse_duration("99999999999999999999999999999m").is_err());
    }

    #[test]
    fn test_parse_timestamp() {
        let eve_format = "2022-04-12T17:20:42.294911-0600";
        parse_timestamp(eve_format).unwrap();

        let eve_format = "2022-04-12T17:20:42.294911-06:00";
        parse_timestamp(eve_format).unwrap();
    }
}
