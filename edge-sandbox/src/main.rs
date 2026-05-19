// ONDC Edge Sandbox — Rust gRPC server for PII redaction via WASM policies.
// Listens on a Unix Domain Socket, receives JSON payloads, runs them through
// a Wasmtime-backed redaction engine, and returns sanitized output.

mod engine;

use std::env;
use tonic::{transport::Server, Request, Response, Status};
use tokio::net::UnixListener;
use tokio_stream::wrappers::UnixListenerStream;
use tracing::{info, warn};

pub mod sandbox {
    tonic::include_proto!("sandbox");
}

use sandbox::sandbox_service_server::{SandboxService, SandboxServiceServer};
use sandbox::{SanitizeRequest, SanitizeResponse};

pub struct SandboxServer {
    redaction_engine: engine::RedactionEngine,
}

#[tonic::async_trait]
impl SandboxService for SandboxServer {
    async fn sanitize_payload(
        &self,
        request: Request<SanitizeRequest>,
    ) -> Result<Response<SanitizeResponse>, Status> {
        let payload = request.into_inner().payload_json;

        match self.redaction_engine.redact(&payload) {
            Ok((sanitized, count)) => {
                info!(fields_redacted = count, "payload sanitized");
                Ok(Response::new(SanitizeResponse {
                    sanitized_json: sanitized,
                    redacted: count > 0,
                    fields_redacted: count,
                }))
            }
            Err(e) => {
                warn!(error = %e, "redaction failed, returning original payload");
                Ok(Response::new(SanitizeResponse {
                    sanitized_json: payload,
                    redacted: false,
                    fields_redacted: 0,
                }))
            }
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt::init();

    let socket_path = env::var("SOCKET_PATH").unwrap_or_else(|_| "/tmp/sandbox.sock".into());

    // Clean up stale socket
    let _ = std::fs::remove_file(&socket_path);

    let engine = engine::RedactionEngine::new()?;
    info!("redaction engine initialized");

    let server = SandboxServer {
        redaction_engine: engine,
    };

    let uds = UnixListener::bind(&socket_path)?;
    let uds_stream = UnixListenerStream::new(uds);

    info!(path = %socket_path, "edge-sandbox listening on UDS");

    Server::builder()
        .add_service(SandboxServiceServer::new(server))
        .serve_with_incoming(uds_stream)
        .await?;

    Ok(())
}
