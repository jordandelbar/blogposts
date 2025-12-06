from fastapi import Request


def read_root(request: Request):
    model = request.app.state.model
    return {
        "message": "Adult Income Prediction API",
        "status": "Model loaded" if model.is_loaded() else "Model not loaded",
        "endpoints": {
            "POST /predict": "Make a prediction",
            "POST /predict/batch": "Make batch predictions",
            "GET /health": "Check API health",
        },
    }
