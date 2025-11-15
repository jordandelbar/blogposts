import pickle

import numpy as np
from loguru import logger

from utils import load_dataframe


def load_model(model_name: str):
    logger.info("Loading model")
    with open(f"./models/{model_name}.pkl", "rb") as handle:
        model = pickle.load(handle)
    logger.info("Returning model")
    return model


def make_prediction(model, x) -> np.ndarray:
    logger.info("Inferring")
    y_predict = model.predict(x)
    logger.info("Inferring done")
    return y_predict


if __name__ == "__main__":
    x, _ = load_dataframe()

    # Dropper/Load approach
    model = load_model("very_safe_model")
    y_predict = make_prediction(model, x)
    print(y_predict)
    # Trojan approach
    model = load_model("another_very_safe_model")
    y_predict = make_prediction(model, x)
    print(y_predict)
