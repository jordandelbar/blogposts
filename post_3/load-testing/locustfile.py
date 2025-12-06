from random import choice, randint

from locust import HttpUser, between, task


class BasePredictionUser(HttpUser):
    abstract = True

    WORKCLASSES = [
        "Private",
        "State-gov",
        "Federal-gov",
        "Self-emp-not-inc",
        "Self-emp-inc",
    ]
    EDUCATIONS = [
        "Bachelors",
        "HS-grad",
        "Masters",
        "Some-college",
        "Assoc-voc",
        "Doctorate",
    ]
    MARITAL_STATUSES = ["Married-civ-spouse", "Never-married", "Divorced", "Separated"]
    OCCUPATIONS = [
        "Tech-support",
        "Craft-repair",
        "Other-service",
        "Sales",
        "Exec-managerial",
        "Prof-specialty",
        "Handlers-cleaners",
        "Machine-op-inspct",
        "Adm-clerical",
    ]
    RELATIONSHIPS = [
        "Wife",
        "Own-child",
        "Husband",
        "Not-in-family",
        "Other-relative",
        "Unmarried",
    ]
    RACES = ["White", "Black", "Asian-Pac-Islander", "Amer-Indian-Eskimo", "Other"]
    SEXES = ["Male", "Female"]
    COUNTRIES = [
        "United-States",
        "Mexico",
        "Philippines",
        "Germany",
        "Canada",
        "India",
        "China",
    ]

    def generate_sample(self) -> dict:
        return {
            "age": float(randint(18, 90)),
            "workclass": choice(self.WORKCLASSES),
            "fnlwgt": float(randint(10000, 500000)),
            "education": choice(self.EDUCATIONS),
            "education_num": float(randint(1, 16)),
            "marital_status": choice(self.MARITAL_STATUSES),
            "occupation": choice(self.OCCUPATIONS),
            "relationship": choice(self.RELATIONSHIPS),
            "race": choice(self.RACES),
            "sex": choice(self.SEXES),
            "capital_gain": float(randint(0, 100000)),
            "capital_loss": float(randint(0, 5000)),
            "hours_per_week": float(randint(1, 99)),
            "native_country": choice(self.COUNTRIES),
        }


class SinglePredictionUser(BasePredictionUser):
    wait_time = between(0.1, 0.5)

    @task
    def predict_single(self):
        """Make a single prediction request"""
        sample = self.generate_sample()
        with self.client.post(
            "/predict", json=sample, catch_response=True, name="/predict (single)"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")


class BatchPredictionUser(BasePredictionUser):
    wait_time = between(1, 3)

    @task(3)
    def predict_batch_small(self):
        batch = [self.generate_sample() for _ in range(10)]
        with self.client.post(
            "/predict/batch",
            json=batch,
            catch_response=True,
            name="/predict/batch (size=10)",
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")

    @task(2)
    def predict_batch_medium(self):
        batch = [self.generate_sample() for _ in range(50)]
        with self.client.post(
            "/predict/batch",
            json=batch,
            catch_response=True,
            name="/predict/batch (size=50)",
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")

    @task(1)
    def predict_batch_large(self):
        batch = [self.generate_sample() for _ in range(100)]
        with self.client.post(
            "/predict/batch",
            json=batch,
            catch_response=True,
            name="/predict/batch (size=100)",
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")


class MixedUser(BasePredictionUser):
    wait_time = between(0.5, 2)

    @task(8)
    def predict_single(self):
        sample = self.generate_sample()
        self.client.post("/predict", json=sample, name="/predict (mixed)")

    @task(2)
    def predict_batch(self):
        batch_size = choice([10, 25, 50])
        batch = [self.generate_sample() for _ in range(batch_size)]
        self.client.post(
            "/predict/batch",
            json=batch,
            name=f"/predict/batch (mixed, size={batch_size})",
        )


class StressTestUser(BasePredictionUser):
    wait_time = between(0, 0.1)

    @task
    def predict_aggressive(self):
        sample = self.generate_sample()
        self.client.post("/predict", json=sample)
