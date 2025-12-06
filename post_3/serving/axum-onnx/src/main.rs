use axum::{
    Router,
    routing::{get, post},
};
use ort_onnx::{
    model::Model,
    routes::{health, predict, predict_batch, root},
};
use std::sync::Arc;
use tower_http::trace::TraceLayer;
use tracing::info;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_target(false)
        .compact()
        .init();

    let model_path = std::env::var("MODEL_PATH").unwrap_or_else(|_| {
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .parent()
            .unwrap()
            .parent()
            .unwrap()
            .join("training")
            .join("models")
            .join("adult_income_model.onnx")
            .to_string_lossy()
            .to_string()
    });

    info!("Loading model from: {}", model_path);

    let num_instances = 4;
    let model = Arc::new(Model::new(&model_path, num_instances).expect("Failed to load model"));

    info!("Model loaded successfully with {} instances", num_instances);

    let app = Router::new()
        .route("/", get(root))
        .route("/health", get(health))
        .route("/predict", post(predict))
        .route("/predict/batch", post(predict_batch))
        .layer(TraceLayer::new_for_http())
        .with_state(model);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8000").await.unwrap();

    info!("Server running on http://0.0.0.0:8000");

    axum::serve(listener, app).await.unwrap();
}
