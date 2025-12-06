use axum::{Json, extract::State, http::StatusCode};
use std::sync::Arc;
use crate::model::{Model, prepare_inputs};
use crate::types::{AdultIncomeInput, PredictionResponse, BatchPredictionResponse};

pub async fn predict(
    State(model): State<Arc<Model>>,
    Json(data): Json<AdultIncomeInput>,
) -> Result<Json<PredictionResponse>, (StatusCode, String)> {
    let inputs = prepare_inputs(&[data]).map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e))?;
    let (labels, probs) = model
        .predict(&inputs)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let prediction = labels[0];
    let probabilities = probs[0..2].to_vec();
    let predicted_income = if prediction == 1 { ">50K" } else { "<=50K" };

    Ok(Json(PredictionResponse {
        prediction,
        probabilities,
        predicted_income: predicted_income.to_string(),
    }))
}

pub async fn predict_batch(
    State(model): State<Arc<Model>>,
    Json(batch): Json<Vec<AdultIncomeInput>>,
) -> Result<Json<BatchPredictionResponse>, (StatusCode, String)> {
    if batch.is_empty() {
        return Err((StatusCode::BAD_REQUEST, "Empty batch".to_string()));
    }

    let inputs = prepare_inputs(&batch).map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e))?;
    let (labels, probs) = model
        .predict(&inputs)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let predictions = labels
        .iter()
        .enumerate()
        .map(|(i, &label)| {
            let probabilities = probs[i * 2..(i + 1) * 2].to_vec();
            let predicted_income = if label == 1 { ">50K" } else { "<=50K" };
            PredictionResponse {
                prediction: label,
                probabilities,
                predicted_income: predicted_income.to_string(),
            }
        })
        .collect();

    Ok(Json(BatchPredictionResponse {
        predictions,
        batch_size: batch.len(),
    }))
}
