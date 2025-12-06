from fastapi import Request


def health_check(request: Request):
    model = request.app.state.model
    is_loaded = model.is_loaded()
    return {
        "status": "healthy" if is_loaded else "unhealthy",
        "model_loaded": is_loaded,
    }
