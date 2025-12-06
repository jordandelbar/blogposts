use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct AdultIncomeInput {
    pub age: f32,
    pub workclass: String,
    pub fnlwgt: f32,
    pub education: String,
    pub education_num: f32,
    pub marital_status: String,
    pub occupation: String,
    pub relationship: String,
    pub race: String,
    pub sex: String,
    pub capital_gain: f32,
    pub capital_loss: f32,
    pub hours_per_week: f32,
    pub native_country: String,
}

#[derive(Debug, Serialize)]
pub struct PredictionResponse {
    pub prediction: i64,
    pub probabilities: Vec<f32>,
    pub predicted_income: String,
}

#[derive(Debug, Serialize)]
pub struct BatchPredictionResponse {
    pub predictions: Vec<PredictionResponse>,
    pub batch_size: usize,
}

#[derive(Debug, Serialize)]
pub struct RootResponse {
    pub message: String,
    pub status: String,
    pub endpoints: EndpointsInfo,
}

#[derive(Debug, Serialize)]
pub struct EndpointsInfo {
    #[serde(rename = "POST /predict")]
    pub post_predict: String,
    #[serde(rename = "GET /health")]
    pub get_health: String,
}

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub model_loaded: bool,
}
