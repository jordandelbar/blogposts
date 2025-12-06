from fastapi import HTTPException, Request

from app.model import prepare_inputs
from app.types import AdultIncomeInput, PredictionResponse


def predict(data: AdultIncomeInput, request: Request) -> PredictionResponse:
    model = request.app.state.model

    if not model.is_loaded():
        raise HTTPException(status_code=503, detail="Model not loaded")

    try:
        input_dict = prepare_inputs([data])
        outputs = model.predict(input_dict)

        # outputs[0] is the predicted class (label)
        # outputs[1] is the probabilities array
        predicted_class = int(outputs[0][0])
        probabilities = outputs[1][0].tolist()

        # Map prediction to income bracket
        income_label = ">50K" if predicted_class == 1 else "<=50K"

        return PredictionResponse(
            prediction=predicted_class,
            probabilities=probabilities,
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
        batch_size = len(data)
        input_dict = prepare_inputs(data)
        outputs = model.predict(input_dict)

        predictions = outputs[0].tolist()
        probabilities = outputs[1].tolist()

        results = []
        for i in range(batch_size):
            income_label = ">50K" if predictions[i] == 1 else "<=50K"
            results.append(
                {
                    "prediction": int(predictions[i]),
                    "probabilities": probabilities[i],
                    "predicted_income": income_label,
                }
            )

        return {"predictions": results, "batch_size": batch_size}

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Batch prediction error: {str(e)}")
