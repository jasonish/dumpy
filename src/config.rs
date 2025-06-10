// SPDX-FileCopyrightText: (C) 2022 Jason Ish <jason@codemonkey.net>
// SPDX-License-Identifier: MIT

use anyhow::{bail, Result};
use serde::Deserialize;
use serde::Serialize;
use std::collections::HashMap;
use tracing::{info, warn};

pub const DEFAULT_PORT: u16 = 7000;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Config {
    pub port: Option<u16>,
    #[serde(default)]
    pub tls: TlsConfig,
    #[serde(default)]
    pub spools: Vec<SpoolConfig>,
    #[serde(default)]
    pub users: HashMap<String, String>,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            port: Some(7000),
            tls: TlsConfig::default(),
            spools: Vec::new(),
            users: HashMap::new(),
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TlsConfig {
    pub enabled: bool,
    pub certificate: String,
    pub key: String,
}

impl Default for TlsConfig {
    fn default() -> Self {
        Self {
            enabled: false,
            certificate: "cert.pem".to_string(),
            key: "key.pem".to_string(),
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SpoolConfig {
    pub name: String,
    pub directory: String,
    pub prefix: Option<String>,
}

impl Config {
    pub fn load() -> Result<Self> {
        if std::fs::metadata("dumpy.yaml").is_ok() {
            let contents = std::fs::read_to_string("dumpy.yaml")?;
            return Ok(serde_yaml::from_str(&contents)?);
        }
        Ok(Self::default())
    }

    pub fn save(&self) -> Result<()> {
        let out = std::fs::File::create("dumpy.yaml")?;
        serde_yaml::to_writer(&out, self)?;
        Ok(())
    }
}

#[derive(Debug, clap::Parser)]
pub struct ConfigCommand {
    #[clap(subcommand)]
    command: SubCommands,
}

#[derive(Debug, clap::Subcommand)]
enum SubCommands {
    /// Set a configuration value.
    #[clap(after_help = "
Configurable parameters:

    port                         Port to liston on
    tls.enabled <true|false>     Enable/disable TLS
    tls.cert <filename>          TLS cert filename
    tls.key <filename>           TLS key filename
")]
    Set { key: String, val: String },
    /// Manage PCAP spool directories
    #[clap(subcommand)]
    Spool(SpoolCommands),
    /// Set user authentication password
    Passwd { username: String, password: String },
}

#[derive(Debug, clap::Subcommand)]
enum SpoolCommands {
    /// List configured PCAP spool directories
    List,
    /// Add a new PCAP spool directory
    Add {
        /// The name of the spool
        name: String,
        /// Directory containing pcap files
        directory: String,
        /// Optional filename prefix
        prefix: Option<String>,
    },
    /// Remove a configured PCAP spool
    Remove { name: String },
}

pub fn config_main(args: ConfigCommand) -> Result<()> {
    tracing_subscriber::fmt()
        .with_writer(std::io::stderr)
        .init();
    let mut config = Config::load()?;
    match args.command {
        SubCommands::Set { key, val } => config_set(&mut config, key, val)?,
        SubCommands::Spool(spool) => match spool {
            SpoolCommands::List => spool_list(&config)?,
            SpoolCommands::Remove { name } => spool_remove(&mut config, &name)?,
            SpoolCommands::Add {
                name,
                directory,
                prefix,
            } => spool_add(&mut config, name, directory, prefix)?,
        },
        SubCommands::Passwd { username, password } => passwd(&mut config, username, password)?,
    }
    Ok(())
}

fn passwd(config: &mut Config, username: String, password: String) -> Result<()> {
    if username.is_empty() {
        bail!("username cannot be empty");
    }
    if password.is_empty() {
        bail!("password cannot be empty");
    }
    let password = bcrypt::hash(&password, bcrypt::DEFAULT_COST)?;
    if config.users.insert(username.clone(), password).is_some() {
        info!("The password has been updated for user {}", &username);
    } else {
        info!("User {} has been created", &username);
    }
    config.save()?;
    Ok(())
}

fn config_set(config: &mut Config, key: String, val: String) -> Result<()> {
    match key.as_ref() {
        "port" => config.port = Some(val.parse()?),
        "tls.enabled" => config.tls.enabled = val.parse()?,
        "tls.cert" => config.tls.certificate = val,
        "tls.key" => config.tls.key = val,
        _ => anyhow::bail!("unknown configuration parameter: {}", &key),
    }
    config.save()?;
    Ok(())
}

fn spool_list(config: &Config) -> Result<()> {
    if config.spools.is_empty() {
        warn!("No configured spools");
    } else {
        for spool in &config.spools {
            println!(
                "- Name={}, Directory={}, Prefix={}",
                &spool.name,
                &spool.directory,
                &spool.prefix.as_ref().unwrap_or(&("<none>".to_string()))
            );
        }
    }
    Ok(())
}

fn spool_remove(config: &mut Config, name: &str) -> Result<()> {
    let mut removed = false;
    config.spools.retain(|s| {
        if s.name == name {
            removed = true;
            false
        } else {
            true
        }
    });
    if !removed {
        warn!("A spool with the name '{}' was not found", name);
    } else {
        config.save()?;
    }
    Ok(())
}

fn spool_add(
    config: &mut Config,
    name: String,
    directory: String,
    prefix: Option<String>,
) -> Result<()> {
    let spool_config = SpoolConfig {
        name,
        directory,
        prefix,
    };
    config.spools.push(spool_config);
    config.save()?;
    Ok(())
}
