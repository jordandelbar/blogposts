use axum::Json;
use crate::types::{RootResponse, EndpointsInfo};

pub async fn root() -> Json<RootResponse> {
    Json(RootResponse {
        message: "Adult Income Prediction API (ONNX Runtime)".to_string(),
        status: "Model loaded".to_string(),
        endpoints: EndpointsInfo {
            post_predict: "Make a prediction".to_string(),
            get_health: "Check API health".to_string(),
        },
    })
}
