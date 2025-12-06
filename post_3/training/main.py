import argparse
from pathlib import Path

import kagglehub
import numpy as np
import optuna
import polars as pl
from skl2onnx import convert_sklearn
from skl2onnx.common.data_types import FloatTensorType
from sklearn.compose import ColumnTransformer
from sklearn.metrics import (
    auc,
    classification_report,
    f1_score,
    fbeta_score,
    precision_recall_curve,
)
from sklearn.model_selection import StratifiedKFold, train_test_split
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, OrdinalEncoder
from xgboost import XGBClassifier


def download_dataset() -> str:
    path = kagglehub.dataset_download("wenruliu/adult-income-dataset")
    return path


def load_csv(path: str) -> pl.DataFrame:
    return pl.read_csv(f"{path}/adult.csv")


def label_target(df: pl.DataFrame) -> pl.DataFrame:
    mapping: dict[str, int] = {"<=50K": 0, ">50K": 1}
    return df.with_columns(income=pl.col("income").replace(mapping).cast(pl.Int64))


def split(
    df: pl.DataFrame,
) -> tuple[pl.DataFrame, pl.DataFrame, pl.DataFrame, pl.DataFrame]:
    X = df.drop("income")
    y = df.select("income")

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.33, stratify=df["income"], random_state=42
    )

    return (
        pl.DataFrame(X_train),
        pl.DataFrame(X_test),
        pl.DataFrame(y_train),
        pl.DataFrame(y_test),
    )


def create_preprocessing_pipeline(X: pl.DataFrame) -> ColumnTransformer:
    # Define ordinal features with proper ordering
    education_order = [
        "Preschool",
        "1st-4th",
        "5th-6th",
        "7th-8th",
        "9th",
        "10th",
        "11th",
        "12th",
        "HS-grad",
        "Some-college",
        "Assoc-voc",
        "Assoc-acdm",
        "Bachelors",
        "Masters",
        "Prof-school",
        "Doctorate",
    ]

    ordinal_features = ["education"]

    all_categorical_cols = [col for col in X.columns if X[col].dtype == pl.Utf8]

    nominal_features = [
        col for col in all_categorical_cols if col not in ordinal_features
    ]
    numerical_cols = [col for col in X.columns if col not in all_categorical_cols]

    transformers = []

    if "education" in X.columns and X["education"].dtype == pl.Utf8:
        transformers.append(
            (
                "ordinal",
                OrdinalEncoder(
                    categories=[education_order],  # type: ignore
                    handle_unknown="use_encoded_value",
                    unknown_value=-1,
                ),
                ["education"],
            )
        )

    if nominal_features:
        transformers.append(
            (
                "nominal",
                OneHotEncoder(handle_unknown="ignore", sparse_output=False),
                nominal_features,
            )
        )

    if numerical_cols:
        transformers.append(("num", "passthrough", numerical_cols))

    preprocessor = ColumnTransformer(transformers=transformers)
    return preprocessor


def train_model(
    X_train: pl.DataFrame, y_train: pl.DataFrame, params: dict | None = None
) -> Pipeline:
    y_train_np = y_train.to_numpy().flatten()

    preprocessor = create_preprocessing_pipeline(X_train)

    pipeline = Pipeline(
        [
            ("preprocessor", preprocessor),
            (
                "classifier",
                XGBClassifier(
                    n_estimators=100,
                    max_depth=5,
                    learning_rate=0.1,
                    scale_pos_weight=3.17,
                    objective="binary:logistic",
                ),
            ),
        ]
    )

    if params:
        pipeline.set_params(**params)

    pipeline.fit(X_train, y_train_np)

    return pipeline


def find_best_threshold(y_true, y_proba, beta=2.0):
    """Find the threshold that maximizes F-beta score"""
    precision, recall, thresholds = precision_recall_curve(y_true, y_proba)
    fbeta_scores = (
        (1 + beta**2) * (precision * recall) / ((beta**2 * precision) + recall)
    )

    # Handle NaN values
    fbeta_scores = np.nan_to_num(fbeta_scores)
    best_idx = np.argmax(fbeta_scores)
    return thresholds[best_idx] if best_idx < len(thresholds) else 0.5


def evaluate_model(X_test: pl.DataFrame, y_test: pl.DataFrame, pipeline: Pipeline):
    y_test_np = y_test.to_numpy().flatten()

    y_proba = pipeline.predict_proba(X_test)[:, 1]

    optimal_threshold = find_best_threshold(y_test_np, y_proba, beta=2.0)
    print(f"Optimal threshold: {optimal_threshold:.3f}")

    y_predict_optimal = (y_proba >= optimal_threshold).astype(int)

    y_predict_default = pipeline.predict(X_test)

    precision, recall, _ = precision_recall_curve(y_test_np, y_proba)
    pr_auc = auc(recall, precision)

    f1_optimal = f1_score(y_test_np, y_predict_optimal)
    fbeta_optimal = fbeta_score(y_test_np, y_predict_optimal, beta=2.0)

    f1_default = f1_score(y_test_np, y_predict_default)
    fbeta_default = fbeta_score(y_test_np, y_predict_default, beta=2.0)

    print(f"PR-AUC Score: {pr_auc:.3f}")
    print(f"\nWith optimal threshold ({optimal_threshold:.3f}):")
    print(f"F1 Score: {f1_optimal:.3f}")
    print(f"F-beta (β=2.0) Score: {fbeta_optimal:.3f}")
    print(classification_report(y_test_np, y_predict_optimal))

    print("\nWith default threshold (0.5):")
    print(f"F1 Score: {f1_default:.3f}")
    print(f"F-beta (β=2.0) Score: {fbeta_default:.3f}")
    print(classification_report(y_test_np, y_predict_default))

    return pr_auc


def objective(trial, X, y):
    params = {
        "classifier__n_estimators": trial.suggest_int("n_estimators", 50, 200),
        "classifier__max_depth": trial.suggest_int("max_depth", 3, 5),
        "classifier__learning_rate": trial.suggest_float("learning_rate", 0.05, 0.1),
        "classifier__subsample": trial.suggest_float("subsample", 0.8, 1.0),
        "classifier__colsample_bytree": trial.suggest_float(
            "colsample_bytree", 0.8, 1.0
        ),
        "classifier__reg_alpha": trial.suggest_float("reg_alpha", 0, 0.1),
        "classifier__reg_lambda": trial.suggest_float("reg_lambda", 0, 0.1),
        "classifier__scale_pos_weight": trial.suggest_float(
            "scale_pos_weight", 2.5, 4.0
        ),
        "classifier__objective": "binary:logistic",
    }

    y_np = y.to_numpy().flatten()

    preprocessor = create_preprocessing_pipeline(X)
    pipeline = Pipeline(
        [("preprocessor", preprocessor), ("classifier", XGBClassifier())]
    )

    pipeline.set_params(**params)

    kf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
    fb_scores = []
    pr_auc_scores = []

    for _, (
        train_idx,
        val_idx,
    ) in enumerate(kf.split(X, y_np)):
        X_train, X_val = (
            X.with_row_index().filter(pl.col("index").is_in(train_idx)).drop("index"),
            X.with_row_index().filter(pl.col("index").is_in(val_idx)).drop("index"),
        )
        y_train, y_val = y_np[train_idx], y_np[val_idx]

        pipeline.fit(X_train, y_train)

        y_proba = pipeline.predict_proba(X_val)[:, 1]
        y_pred = pipeline.predict(X_val)

        fbeta = fbeta_score(y_val, y_pred, beta=2.0, pos_label=1)
        precision, recall, _ = precision_recall_curve(y_val, y_proba)
        pr_auc = auc(recall, precision)
        fb_scores.append(fbeta)
        pr_auc_scores.append(pr_auc)

    mean_fbeta = float(np.mean(fb_scores))
    mean_pr_auc = float(np.mean(pr_auc_scores))

    return mean_fbeta, mean_pr_auc


def serialize_to_onnx(
    pipeline: Pipeline, X_sample: pl.DataFrame, model_path: str = "model.onnx"
) -> str:
    from onnxmltools.convert.xgboost.operator_converters.XGBoost import convert_xgboost
    from skl2onnx import update_registered_converter
    from skl2onnx.common.data_types import StringTensorType
    from skl2onnx.common.shape_calculator import (
        calculate_linear_classifier_output_shapes,
    )
    from xgboost import XGBClassifier

    update_registered_converter(
        XGBClassifier,
        "XGBoostXGBClassifier",
        calculate_linear_classifier_output_shapes,
        convert_xgboost,
        options={"nocl": [True, False], "zipmap": [True, False, "columns"]},
    )

    initial_types = []
    for col in X_sample.columns:
        if X_sample[col].dtype == pl.Utf8:
            initial_types.append((col, StringTensorType([None, 1])))
        else:
            initial_types.append((col, FloatTensorType([None, 1])))

    onnx_model = convert_sklearn(
        pipeline,
        initial_types=initial_types,
        target_opset={"": 12, "ai.onnx.ml": 2},
        options={XGBClassifier: {"zipmap": False, "nocl": False}},
    )  # type: ignore

    output_path = Path(model_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, "wb") as f:
        f.write(onnx_model.SerializeToString())  # type: ignore

    return str(output_path)


def serialize_to_pickle(pipeline: Pipeline, model_path: str = "model.pkl") -> str:
    from joblib import dump

    output_path = Path(model_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    dump(pipeline, output_path)
    return str(output_path)


def main():
    parser = argparse.ArgumentParser(
        description="Train XGBoost model on Adult Income dataset and export to ONNX"
    )
    parser.add_argument(
        "--n-trials",
        type=int,
        default=100,
        help="Number of Optuna trials for hyperparameter optimization (default: 100)",
    )
    args = parser.parse_args()

    path = download_dataset()
    df = load_csv(path)
    print(f"Dataset shape: {df.shape}")

    df = label_target(df)

    X_train, X_test, y_train, y_test = split(df)

    print(f"\nStarting hyperparameter optimization with {args.n_trials} trials...")
    study = optuna.create_study(directions=["maximize", "maximize"])
    study.optimize(
        lambda trial: objective(trial, X_train, y_train),
        n_trials=args.n_trials,
        show_progress_bar=True,
    )

    pareto_trials = study.best_trials

    print(f"\nFound {len(pareto_trials)} trials on the Pareto front.")

    min_pr_auc_threshold = 0.75

    best_fbeta_trial = None
    max_fbeta_score = -1.0

    for t in pareto_trials:
        fbeta_value = t.values[0]
        auc_score = t.values[1]

        if auc_score >= min_pr_auc_threshold and fbeta_value > max_fbeta_score:
            max_fbeta_score = fbeta_value
            best_fbeta_trial = t

    if best_fbeta_trial is None:
        print(
            "No trial met the minimum AUC threshold. Consider lowering the threshold."
        )
    else:
        print(
            f"Best trial (F2 with min AUC): F2={best_fbeta_trial.values[0]:.3f}, AUC={best_fbeta_trial.values[1]:.3f}"
        )

        best_params = best_fbeta_trial.params

        pipeline_params = {f"classifier__{k}": v for k, v in best_params.items()}
        final_pipeline = train_model(X_train, y_train, pipeline_params)

        print("\nFinal model performance on test set:")
        _ = evaluate_model(X_test, y_test, final_pipeline)

        print("\nSerializing model to ONNX format...")
        onnx_path = serialize_to_onnx(
            final_pipeline, X_train, "models/adult_income_model.onnx"
        )
        print(f"ONNX model saved to: {onnx_path}")

        pickle_path = serialize_to_pickle(
            final_pipeline, "models/adult_income_model.pkl"
        )
        print(f"Pickle model saved to: {pickle_path}")


if __name__ == "__main__":
    main()
