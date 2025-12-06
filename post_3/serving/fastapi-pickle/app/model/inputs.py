import polars as pl

from app.types import AdultIncomeInput


def prepare_dataframe(data: list[AdultIncomeInput]) -> pl.DataFrame:
    return pl.DataFrame(
        {
            "age": [d.age for d in data],
            "workclass": [d.workclass for d in data],
            "fnlwgt": [d.fnlwgt for d in data],
            "education": [d.education for d in data],
            "educational-num": [d.education_num for d in data],
            "marital-status": [d.marital_status for d in data],
            "occupation": [d.occupation for d in data],
            "relationship": [d.relationship for d in data],
            "race": [d.race for d in data],
            "gender": [d.sex for d in data],
            "capital-gain": [d.capital_gain for d in data],
            "capital-loss": [d.capital_loss for d in data],
            "hours-per-week": [d.hours_per_week for d in data],
            "native-country": [d.native_country for d in data],
        }
    )
