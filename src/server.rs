// SPDX-FileCopyrightText: (C) 2022 Jason Ish <jason@codemonkey.net>
// SPDX-License-Identifier: MIT

use crate::config::{Config, DEFAULT_PORT};
use anyhow::Result;
use axum::http::Request;
use axum::http::{HeaderMap, HeaderValue, StatusCode, Uri};
use axum::response::IntoResponse;
use axum::Extension;
use axum_server::tls_rustls::RustlsConfig;
use base64::engine::general_purpose::STANDARD as BASE64;
use base64::engine::Engine;
use rust_embed::RustEmbed;
use serde::Deserialize;
use std::collections::HashMap;
use std::net::SocketAddr;
use tracing::{error, info};

#[derive(RustEmbed)]
#[folder = "www/"]
struct Asset;

#[derive(Deserialize, Debug)]
#[allow(dead_code)]
struct FetchRequest {
    #[serde(rename = "default-timezone-offset")]
    default_timezone_offset: String,
    #[serde(rename = "query-type")]
    query_type: String,
    filter: String,
    #[serde(rename = "start-time")]
    start_time: String,
    duration: String,
    event: Option<String>,
    #[serde(rename = "duration-before")]
    duration_before: String,
    #[serde(rename = "duration-after")]
    duration_after: String,
}

#[tokio::main]
pub async fn start_server() -> Result<()> {
    crate::logging::init_logging(tracing::Level::INFO, false);

    tokio::spawn(async move {
        tokio::signal::ctrl_c()
            .await
            .expect("Failed to register CTRL-C handler");
        std::process::exit(0);
    });

    let config = crate::config::Config::load()?;
    let app = axum::Router::new()
        .route("/api/spools", axum::routing::get(get_spools))
        .route("/fetch", axum::routing::post(crate::fetch::fetch_post))
        .route("/fetch", axum::routing::get(crate::fetch::fetch_get))
        .fallback(axum::routing::get(fallback_handler))
        .layer(Extension(config.clone()))
        .layer(axum::middleware::from_fn(move |request, next| {
            let authenticator = Authenticator {
                users: config.users.clone(),
            };
            authenticator.authenticate(request, next)
        }));
    let addr: SocketAddr = format!("[::]:{}", config.port.unwrap_or(DEFAULT_PORT))
        .parse()
        .unwrap();
    info!("Starting server on {}, TLS={}", &addr, config.tls.enabled);

    let service = app.into_make_service();

    let bind = if config.tls.enabled {
        let tls_config =
            match RustlsConfig::from_pem_file(config.tls.certificate, config.tls.key).await {
                Ok(config) => config,
                Err(err) => {
                    error!("Failed to load TLS certificator and/or key: {:?}", err);
                    std::process::exit(1);
                }
            };
        axum_server::bind_rustls(addr, tls_config)
            .serve(service)
            .await
    } else {
        axum_server::bind(addr).serve(service).await
    };
    if let Err(err) = bind {
        error!("Failed to start server: error={:?}", err);
        std::process::exit(1);
    }
    Ok(())
}

async fn get_spools(Extension(config): Extension<Config>) -> impl IntoResponse {
    let spools: Vec<String> = config.spools.iter().map(|s| s.name.to_string()).collect();
    axum::Json(spools)
}

async fn fallback_handler(uri: Uri) -> impl IntoResponse {
    let path = if uri.path() == "/" {
        "index.html"
    } else {
        uri.path().trim_start_matches('/')
    };

    match Asset::get(path) {
        None => {
            let response = serde_json::json!({
                "error": "no resource at path",
                "path": &path,
            });
            (StatusCode::NOT_FOUND, axum::Json(response)).into_response()
        }
        Some(body) => {
            let data = body.data.into_owned();
            let mime = mime_guess::from_path(path).first_or_octet_stream();
            let mut headers = HeaderMap::new();
            headers.insert(
                axum::http::header::CONTENT_TYPE,
                HeaderValue::from_str(mime.as_ref()).unwrap(),
            );
            (StatusCode::OK, headers, data).into_response()
        }
    }
}

#[derive(Clone)]
pub struct Authenticator {
    pub users: HashMap<String, String>,
}

impl Authenticator {
    fn decode_username_password<B>(request: &Request<B>) -> Option<(String, String)> {
        let header = request
            .headers()
            .get("Authorization")
            .and_then(|h| h.to_str().ok())?;

        if header.starts_with("Basic ") {
            if let Some(encoded) = header.split(' ').nth(1) {
                if let Ok(usernamepassword) = BASE64.decode(encoded) {
                    if let Ok(usernamepassword) = String::from_utf8(usernamepassword) {
                        let parts: Vec<&str> = usernamepassword.splitn(2, ':').collect();
                        if parts.len() == 2 {
                            return Some((parts[0].to_string(), parts[1].to_string()));
                        }
                    }
                }
            }
        }

        None
    }

    async fn authenticate(
        self,
        request: axum::extract::Request,
        next: axum::middleware::Next,
    ) -> impl IntoResponse {
        if self.users.is_empty() {
            return next.run(request).await;
        }

        if let Some((username, password)) = Self::decode_username_password(&request) {
            if let Some(hashed) = self.users.get(&username) {
                if bcrypt::verify(password, hashed).unwrap_or(false) {
                    return next.run(request).await;
                }
            }
        }

        let mut headers = HeaderMap::new();
        headers.insert(
            axum::http::header::WWW_AUTHENTICATE,
            HeaderValue::from_str("Basic real=restricted").unwrap(),
        );

        (StatusCode::UNAUTHORIZED, headers, "Unauthorized").into_response()
    }
}
