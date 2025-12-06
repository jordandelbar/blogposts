from pydantic import BaseModel


class AdultIncomeInput(BaseModel):
    age: float
    workclass: str
    fnlwgt: float
    education: str
    education_num: float
    marital_status: str
    occupation: str
    relationship: str
    race: str
    sex: str
    capital_gain: float
    capital_loss: float
    hours_per_week: float
    native_country: str

    class Config:
        json_schema_extra = {
            "example": {
                "age": 39.0,
                "workclass": "State-gov",
                "fnlwgt": 77516.0,
                "education": "Bachelors",
                "education_num": 13.0,
                "marital_status": "Never-married",
                "occupation": "Adm-clerical",
                "relationship": "Not-in-family",
                "race": "White",
                "sex": "Male",
                "capital_gain": 2174.0,
                "capital_loss": 0.0,
                "hours_per_week": 40.0,
                "native_country": "United-States",
            }
        }


class PredictionResponse(BaseModel):
    prediction: int
    probabilities: list[float]
    predicted_income: str
