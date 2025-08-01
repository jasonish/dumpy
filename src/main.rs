// SPDX-FileCopyrightText: (C) 2022 Jason Ish <jason@codemonkey.net>
// SPDX-License-Identifier: MIT

mod config;
mod export;
mod fetch;
mod logging;
mod purge;
mod server;

use crate::export::ExportArgs;
use clap::builder::styling::{AnsiColor, Effects, Styles};
use clap::Parser;
use tracing::error;

const STYLES: Styles = Styles::styled()
    .header(AnsiColor::Green.on_default().effects(Effects::BOLD))
    .usage(AnsiColor::Green.on_default().effects(Effects::BOLD))
    .literal(AnsiColor::Cyan.on_default().effects(Effects::BOLD))
    .placeholder(AnsiColor::BrightCyan.on_default())
    .error(AnsiColor::Red.on_default().effects(Effects::BOLD))
    .valid(AnsiColor::Green.on_default().effects(Effects::BOLD))
    .invalid(AnsiColor::Red.on_default().effects(Effects::BOLD));

/// PCAP spool directory processor
#[derive(Parser, Debug)]
#[clap(version, max_term_width = 80, about, long_about = None, color = clap::ColorChoice::Auto, styles = STYLES)]
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
    /// Start the web server for PCAP viewing
    Server {
        /// Spool directory containing PCAP files
        #[clap(long, short)]
        directory: Option<String>,

        /// Name of the spool (default: "default")
        #[clap(long, short)]
        name: Option<String>,

        /// Optional filename prefix filter
        #[clap(long, short)]
        prefix: Option<String>,

        /// Port to listen on
        #[clap(long)]
        port: Option<u16>,
    },
    /// Configure Dumpy settings and PCAP spools
    Config(config::ConfigCommand),
    /// Purge old PCAP files from spools
    Purge(purge::PurgeArgs),
}

fn main() {
    // Initialize the timezone offset before any threads might be created.
    logging::init_offset();

    let args = Args::parse();
    if let Err(err) = match args.command {
        Commands::Export(sub_args) => export::main(sub_args),
        Commands::Server {
            directory,
            name,
            prefix,
            port,
        } => server::start_server(directory, name, prefix, port),
        Commands::Config(args) => config::config_main(args),
        Commands::Purge(args) => purge::purge_main(args),
    } {
        error!("command failed: {:?}", err);
    }
}
