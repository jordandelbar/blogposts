from fastapi import HTTPException, Request

from app.model import prepare_dataframe
from app.types import AdultIncomeInput, PredictionResponse


def predict(data: AdultIncomeInput, request: Request) -> PredictionResponse:
    model = request.app.state.model

    if not model.is_loaded():
        raise HTTPException(status_code=503, detail="Model not loaded")

    try:
        input_df = prepare_dataframe([data])

        predictions, probabilities = model.predict(input_df)

        predicted_class = int(predictions[0])
        probs = probabilities[0].tolist()

        income_label = ">50K" if predicted_class == 1 else "<=50K"

        return PredictionResponse(
            prediction=predicted_class,
            probabilities=probs,
            predicted_income=income_label,
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction error: {str(e)}")


def predict_batch(data: list[AdultIncomeInput], request: Request):
    model = request.app.state.model

    if not model.is_loaded():
        raise HTTPException(status_code=503, detail="Model not loaded")

    if len(data) == 0:
        raise HTTPException(status_code=400, detail="Empty batch")

    try:
        input_df = prepare_dataframe(data)

        predictions, probabilities = model.predict(input_df)

        results = []
        for i in range(len(data)):
            income_label = ">50K" if predictions[i] == 1 else "<=50K"
            results.append(
                {
                    "prediction": int(predictions[i]),
                    "probabilities": probabilities[i].tolist(),
                    "predicted_income": income_label,
                }
            )

        return {"predictions": results, "batch_size": len(data)}

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Batch prediction error: {str(e)}")
