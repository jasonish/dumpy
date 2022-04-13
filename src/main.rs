// SPDX-License-Identifier: MIT
//
// Copyright (C) 2022 Jason Ish

mod config;
mod export;
mod fetch;
mod server;

use crate::export::ExportArgs;
use clap::Parser;
use tracing::error;

#[derive(Parser, Debug)]
#[clap(version, max_term_width = 80)]
struct Args {
    #[clap(subcommand)]
    subcommands: Commands,
}

#[derive(clap::Parser, Debug)]
enum Commands {
    Export(ExportArgs),
    Server,
    Config(crate::config::ConfigCommand),
}

fn main() {
    let args = Args::parse();
    if let Err(err) = match args.subcommands {
        Commands::Export(sub_args) => export::main(sub_args),
        Commands::Server => server::start_server(),
        Commands::Config(args) => crate::config::config_main(args),
    } {
        error!("command failed: {:?}", err);
    }
}
