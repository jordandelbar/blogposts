import logging
from pathlib import Path

from joblib import load
from sklearn.pipeline import Pipeline

logger = logging.getLogger(__name__)


class Model:
    def __init__(self, model_path: Path):
        self.model_path = model_path
        self.pipeline: Pipeline | None = None
        self._load_model()

    def _load_model(self):
        try:
            self.pipeline = load(str(self.model_path))
            logger.info(f"Model loaded successfully from {self.model_path}")
        except Exception as e:
            logger.error(f"Error loading model: {e}")
            self.pipeline = None

    def is_loaded(self) -> bool:
        return self.pipeline is not None

    def predict(self, input_df):
        if self.pipeline is None:
            raise RuntimeError("Model not loaded")

        predictions = self.pipeline.predict(input_df)
        probabilities = self.pipeline.predict_proba(input_df)
        return predictions, probabilities
