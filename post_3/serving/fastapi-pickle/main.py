import logging
from pathlib import Path

from fastapi import FastAPI, Request

from app.model import Model
from app.routes import health_check, predict, predict_batch, read_root
from app.types import AdultIncomeInput, PredictionResponse

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Adult Income Prediction API (Pickle)")

MODEL_PATH = Path("/app/models/adult_income_model.pkl")

model = Model(MODEL_PATH)
app.state.model = model


@app.get("/")
def root(request: Request):
    return read_root(request)


@app.get("/health")
def health(request: Request):
    return health_check(request)


@app.post("/predict", response_model=PredictionResponse)
def single_predict(data: AdultIncomeInput, request: Request):
    return predict(data, request)


@app.post("/predict/batch")
def batch_predict(data: list[AdultIncomeInput], request: Request):
    return predict_batch(data, request)
