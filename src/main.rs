// SPDX-License-Identifier: MIT
//
// Copyright (C) 2022 Jason Ish

mod config;
mod export;
mod fetch;
mod logging;
mod server;

use crate::export::ExportArgs;
use clap::Parser;
use tracing::error;

/// PCAP spool directory processor
#[derive(Parser, Debug)]
#[clap(version, max_term_width = 80, about, long_about = None)]
struct Args {
    /// Enable more verbose logging
    #[clap(long, short, global = true, action = clap::ArgAction::Count)]
    verbose: u8,

    #[clap(subcommand)]
    command: Commands,
}

#[derive(clap::Parser, Debug)]
enum Commands {
    /// Export data from PCAP files
    Export(ExportArgs),
    Server,
    Config(config::ConfigCommand),
}

fn main() {
    // Initialize the timezone offset before any threads might be created.
    logging::init_offset();

    let args = Args::parse();
    if let Err(err) = match args.command {
        Commands::Export(sub_args) => export::main(sub_args),
        Commands::Server => server::start_server(),
        Commands::Config(args) => config::config_main(args),
    } {
        error!("command failed: {:?}", err);
    }
}
