import numpy as np

from typing import Any

from sklearn.datasets import load_iris


def load_dataframe() -> tuple[np.ndarray, np.ndarray]:
    iris: dict[str, Any] = load_iris()
    x = iris["data"]
    y = iris["target"]
    return x, y
