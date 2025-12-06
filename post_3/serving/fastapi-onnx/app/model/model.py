import logging
from pathlib import Path

import onnxruntime as ort

logger = logging.getLogger(__name__)


class Model:
    def __init__(self, model_path: Path):
        self.model_path = model_path
        self.session = None
        self._load_model()

    def _load_model(self):
        try:
            self.session = ort.InferenceSession(str(self.model_path))
            logger.info(f"Model loaded successfully from {self.model_path}")
            input_names = [inp.name for inp in self.session.get_inputs()]
            logger.info(f"Model expects inputs: {input_names}")

        except Exception as e:
            logger.error(f"Error loading model: {e}")
            self.session = None

    def is_loaded(self) -> bool:
        return self.session is not None

    def predict(self, input_dict: dict):
        if self.session is None:
            raise RuntimeError("Model not loaded")

        return self.session.run(None, input_dict)
