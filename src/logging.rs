// SPDX-License-Identifier: MIT
//
// Copyright (C) 2022 Jason Ish

use time::macros::format_description;
use time::UtcOffset;
use tracing::Level;
use tracing_subscriber::fmt::time::OffsetTime;

static mut OFFSET: Option<UtcOffset> = None;

pub fn init_offset() {
    if let Ok(offset) = UtcOffset::current_local_offset() {
        unsafe {
            OFFSET = Some(offset);
        }
    }
}

pub fn init_logging(level: Level, json: bool) {
    let (offset, format) = if let Some(offset) = unsafe { OFFSET } {
        (
            offset,
            format_description!("[year]-[month]-[day] [hour]:[minute]:[second]"),
        )
    } else {
        (
            UtcOffset::UTC,
            format_description!("[year]-[month]-[day] [hour]:[minute]:[second]Z"),
        )
    };

    let timer = OffsetTime::new(offset, format);
    let logger = tracing_subscriber::fmt()
        .with_max_level(level)
        .with_timer(timer)
        .with_writer(std::io::stderr);
    if json {
        logger.json().init();
    } else {
        logger.init();
    }
}
