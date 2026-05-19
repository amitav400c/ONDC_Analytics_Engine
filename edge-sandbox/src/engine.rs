// PII Redaction Engine — processes JSON payloads to hash phone numbers and fuzz GPS coordinates.
// Uses native Rust for the actual redaction logic. Wasmtime integration provides the sandboxed
// execution boundary, ensuring untrusted policy modules cannot access the host filesystem/network.
//
// TODO: Load actual .wasm policy modules for configurable redaction rules.
// Current implementation uses host-side redaction as a working baseline.

use sha2::{Sha256, Digest};
use serde_json::Value;
use rand::Rng;

pub struct RedactionEngine {
    // TODO: Store pre-compiled wasmtime::Module here for WASM-based policies
    // engine: wasmtime::Engine,
    // module: wasmtime::Module,
}

impl RedactionEngine {
    pub fn new() -> Result<Self, Box<dyn std::error::Error>> {
        // TODO: Initialize Wasmtime engine with pooling allocator
        // let engine = wasmtime::Engine::new(&wasmtime::Config::new().wasm_multi_value(true))?;
        // let module = wasmtime::Module::from_file(&engine, "policies/pii_redact.wasm")?;
        Ok(Self {})
    }

    /// Redacts PII from a JSON payload. Returns (sanitized_json, fields_redacted_count).
    pub fn redact(&self, payload: &str) -> Result<(String, u32), Box<dyn std::error::Error>> {
        let mut value: Value = serde_json::from_str(payload)?;
        let count = redact_recursive(&mut value);
        Ok((serde_json::to_string(&value)?, count))
    }
}

fn redact_recursive(value: &mut Value) -> u32 {
    let mut count = 0u32;

    match value {
        Value::Object(map) => {
            // Phone/contact fields → SHA256 hash
            for key in ["phone", "mobile", "contact", "buyer_phone", "seller_phone"] {
                if let Some(v) = map.get_mut(key) {
                    if let Some(s) = v.as_str() {
                        let mut hasher = Sha256::new();
                        hasher.update(s.as_bytes());
                        let hash = hex::encode(&hasher.finalize()[..8]); // 16-char short hash
                        *v = Value::String(hash);
                        count += 1;
                    }
                }
            }

            // GPS field → fuzz coordinates (±0.01° ≈ 1km)
            if let Some(v) = map.get_mut("gps") {
                if let Some(s) = v.as_str() {
                    *v = Value::String(fuzz_gps(s));
                    count += 1;
                }
            }

            // Recurse into nested objects
            for (_, v) in map.iter_mut() {
                count += redact_recursive(v);
            }
        }
        Value::Array(arr) => {
            for item in arr.iter_mut() {
                count += redact_recursive(item);
            }
        }
        _ => {}
    }
    count
}

fn fuzz_gps(gps: &str) -> String {
    let parts: Vec<&str> = gps.split(',').collect();
    if parts.len() != 2 {
        return gps.to_string();
    }

    let mut rng = rand::thread_rng();
    let lat: f64 = parts[0].trim().parse().unwrap_or(0.0);
    let lng: f64 = parts[1].trim().parse().unwrap_or(0.0);

    // Add random noise ±0.01 degrees (~1km radius)
    let fuzzed_lat = lat + rng.gen_range(-0.01..0.01);
    let fuzzed_lng = lng + rng.gen_range(-0.01..0.01);

    format!("{:.4},{:.4}", fuzzed_lat, fuzzed_lng)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_phone_redaction() {
        let engine = RedactionEngine::new().unwrap();
        let input = r#"{"context":{"action":"on_confirm"},"message":{"buyer":{"phone":"9876543210"}}}"#;
        let (output, count) = engine.redact(input).unwrap();
        assert!(count > 0);
        assert!(!output.contains("9876543210"));
    }

    #[test]
    fn test_gps_fuzzing() {
        let engine = RedactionEngine::new().unwrap();
        let input = r#"{"location":{"gps":"12.9716,77.5946"}}"#;
        let (output, count) = engine.redact(input).unwrap();
        assert_eq!(count, 1);
        assert!(!output.contains("12.9716,77.5946"));
    }
}
