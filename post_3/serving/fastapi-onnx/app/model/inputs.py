import numpy as np

from app.types import AdultIncomeInput


def prepare_inputs(data: list[AdultIncomeInput]) -> dict:
    input_dict = {
        "age": np.array([[d.age] for d in data], dtype=np.float32),
        "workclass": np.array([[d.workclass] for d in data], dtype=object),
        "fnlwgt": np.array([[d.fnlwgt] for d in data], dtype=np.float32),
        "education": np.array([[d.education] for d in data], dtype=object),
        "educational_num": np.array(
            [[d.education_num] for d in data], dtype=np.float32
        ),
        "marital_status": np.array([[d.marital_status] for d in data], dtype=object),
        "occupation": np.array([[d.occupation] for d in data], dtype=object),
        "relationship": np.array([[d.relationship] for d in data], dtype=object),
        "race": np.array([[d.race] for d in data], dtype=object),
        "gender": np.array([[d.sex] for d in data], dtype=object),
        "capital_gain": np.array([[d.capital_gain] for d in data], dtype=np.float32),
        "capital_loss": np.array([[d.capital_loss] for d in data], dtype=np.float32),
        "hours_per_week": np.array(
            [[d.hours_per_week] for d in data], dtype=np.float32
        ),
        "native_country": np.array([[d.native_country] for d in data], dtype=object),
    }

    return input_dict
