use axum::Json;
use crate::types::HealthResponse;

pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        model_loaded: true,
    })
}
