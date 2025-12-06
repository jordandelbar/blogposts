import argparse
import time
from typing import Any

import httpx
import kagglehub
import numpy as np
import polars as pl
from joblib import load
from sklearn.pipeline import Pipeline
from tabulate import tabulate


def download_dataset() -> str:
    path = kagglehub.dataset_download("wenruliu/adult-income-dataset")
    return path


def load_csv(path: str) -> pl.DataFrame:
    return pl.read_csv(f"{path}/adult.csv")


def label_target(df: pl.DataFrame) -> pl.DataFrame:
    mapping: dict[str, int] = {"<=50K": 0, ">50K": 1}
    return df.with_columns(income=pl.col("income").replace(mapping).cast(pl.Int64))


def load_trained_pipeline(model_path: str) -> Pipeline:
    return load(model_path)


def prepare_row_for_api(row: dict[str, Any]) -> dict[str, Any]:
    return {
        "age": float(row["age"]),
        "workclass": str(row["workclass"]),
        "fnlwgt": float(row["fnlwgt"]),
        "education": str(row["education"]),
        "education_num": float(row["educational-num"]),
        "marital_status": str(row["marital-status"]),
        "occupation": str(row["occupation"]),
        "relationship": str(row["relationship"]),
        "race": str(row["race"]),
        "sex": str(row["gender"]),
        "capital_gain": float(row["capital-gain"]),
        "capital_loss": float(row["capital-loss"]),
        "hours_per_week": float(row["hours-per-week"]),
        "native_country": str(row["native-country"]),
    }


def get_training_predictions(
    pipeline: Pipeline, X: pl.DataFrame
) -> tuple[np.ndarray, np.ndarray]:
    predictions = pipeline.predict(X)
    probabilities = pipeline.predict_proba(X)
    return predictions, probabilities


def query_service(
    url: str, data: list[dict[str, Any]], timeout: float = 30.0
) -> tuple[list[int], list[list[float]], float]:
    start_time = time.time()

    try:
        response = httpx.post(
            url,
            json=data,
            timeout=timeout,
            headers={"Content-Type": "application/json"},
        )
        response.raise_for_status()

        elapsed = time.time() - start_time
        result = response.json()

        if isinstance(result, dict) and "predictions" in result:
            if (
                isinstance(result["predictions"], list)
                and len(result["predictions"]) > 0
            ):
                if isinstance(result["predictions"][0], dict):
                    predictions = [p["prediction"] for p in result["predictions"]]
                    probabilities = [p["probabilities"] for p in result["predictions"]]
                else:
                    predictions = result["predictions"]
                    probabilities = result["probabilities"]
            else:
                raise ValueError(f"Empty predictions list: {result}")
        elif isinstance(result, dict) and "prediction" in result:
            predictions = [result["prediction"]]
            probabilities = [result["probabilities"]]
        else:
            raise ValueError(f"Unexpected response format: {result}")

        return predictions, probabilities, elapsed

    except Exception as e:
        print(f"Error querying {url}: {e}")
        raise


def calculate_metrics(
    truth_preds: np.ndarray,
    truth_probs: np.ndarray,
    service_preds: list[int],
    service_probs: list[list[float]],
    tolerance: float = 1e-6,
) -> dict[str, float | int | bool]:
    """
    Calculate skew metrics between training and serving predictions

    Args:
        truth_preds: Ground truth predictions from training pipeline
        truth_probs: Ground truth probabilities from training pipeline
        service_preds: Predictions from serving endpoint
        service_probs: Probabilities from serving endpoint
        tolerance: Tolerance level for probability comparison (default: 1e-6)

    Returns:
        dict with metrics: prediction_match_rate, mae_prob, max_prob_diff,
                          prediction_mismatches, prob_within_tolerance
    """
    service_preds_arr = np.array(service_preds)
    service_probs_arr = np.array(service_probs)

    pred_matches = (truth_preds == service_preds_arr).sum()
    prediction_match_rate = pred_matches / len(truth_preds)

    prob_diffs = np.abs(truth_probs - service_probs_arr)
    mae_prob = np.mean(prob_diffs)
    max_prob_diff = np.max(prob_diffs)

    within_tolerance = np.all(prob_diffs < tolerance)
    samples_within_tolerance = np.sum(np.all(prob_diffs < tolerance, axis=1))
    prob_match_rate = samples_within_tolerance / len(truth_preds)

    prediction_mismatches = len(truth_preds) - pred_matches

    return {
        "prediction_match_rate": float(prediction_match_rate),
        "mae_prob": float(mae_prob),
        "max_prob_diff": float(max_prob_diff),
        "prediction_mismatches": int(prediction_mismatches),
        "prob_within_tolerance": bool(within_tolerance),
        "prob_match_rate": float(prob_match_rate),
        "tolerance": float(tolerance),
    }


def format_results_table(results: dict[str, dict], tolerance: float = 1e-6) -> str:
    headers = [
        "Service",
        "Pred Match %",
        "Prob Match %",
        "Max Diff",
        "MAE",
        "Time (s)",
    ]

    rows = []
    for service_name, metrics in results.items():
        rows.append(
            [
                service_name,
                f"{metrics['prediction_match_rate'] * 100:.2f}%",
                f"{metrics['prob_match_rate'] * 100:.2f}%",
                f"{metrics['max_prob_diff']:.2e}",
                f"{metrics['mae_prob']:.2e}",
                f"{metrics.get('avg_time', 0):.4f}",
            ]
        )

    return tabulate(rows, headers=headers, tablefmt="grid")


def main():
    parser = argparse.ArgumentParser(
        description="Analyze prediction skew between training and serving"
    )
    parser.add_argument(
        "--model-path",
        type=str,
        default="models/adult_income_model.pkl",
        help="Path to the trained pickle model",
    )
    parser.add_argument(
        "--sample-size",
        type=int,
        default=100,
        help="Number of samples to test (default: 100)",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=10,
        help="Batch size for API requests (default: 10)",
    )
    parser.add_argument(
        "--pickle-url",
        type=str,
        default="http://localhost:8000/predict/batch",
        help="FastAPI+Pickle service URL",
    )
    parser.add_argument(
        "--onnx-fastapi-url",
        type=str,
        default="http://localhost:8001/predict/batch",
        help="FastAPI+ONNX service URL",
    )
    parser.add_argument(
        "--onnx-axum-url",
        type=str,
        default="http://localhost:3000/predict/batch",
        help="Axum+ONNX service URL",
    )

    args = parser.parse_args()

    print("=" * 80)
    print("ONNX Serving Skew Analysis")
    print("=" * 80)
    print()

    print("Loading dataset...")
    path = download_dataset()
    df = load_csv(path)
    df = label_target(df)

    test_df = df.sample(n=args.sample_size, seed=42)
    X_test = test_df.drop("income")

    print(f"Loaded {len(test_df)} test samples")
    print()

    print(f"Loading trained pipeline from {args.model_path}...")
    pipeline = load_trained_pipeline(args.model_path)
    print("Pipeline loaded successfully")
    print()

    print("Getting ground truth predictions from training pipeline...")
    truth_preds, truth_probs = get_training_predictions(pipeline, X_test)
    print(f"Truth predictions shape: {truth_preds.shape}")
    print(f"Truth probabilities shape: {truth_probs.shape}")
    print()

    rows = X_test.to_dicts()
    api_data = [prepare_row_for_api(row) for row in rows]

    results = {}

    services = [
        ("FastAPI + Pickle", args.pickle_url),
        ("FastAPI + ONNX", args.onnx_fastapi_url),
        ("Axum + ONNX", args.onnx_axum_url),
    ]

    for service_name, service_url in services:
        print(f"Testing {service_name}...")
        print(f"  URL: {service_url}")

        try:
            all_preds = []
            all_probs = []
            total_time = 0.0
            num_batches = (len(api_data) + args.batch_size - 1) // args.batch_size

            for i in range(0, len(api_data), args.batch_size):
                batch = api_data[i : i + args.batch_size]
                preds, probs, elapsed = query_service(service_url, batch)
                all_preds.extend(preds)
                all_probs.extend(probs)
                total_time += elapsed

            # Calculate metrics with 1e-6 tolerance
            metrics = calculate_metrics(
                truth_preds, truth_probs, all_preds, all_probs, tolerance=1e-6
            )
            metrics["avg_time"] = total_time / num_batches

            results[service_name] = metrics

            print("Success")
            print(
                f"    - Prediction Match: {metrics['prediction_match_rate'] * 100:.2f}%"
            )
            print(
                f"    - Probability Match (< 1e-6): {metrics['prob_match_rate'] * 100:.2f}%"
            )
            print(f"    - Max Probability Diff: {metrics['max_prob_diff']:.2e}")
            print(f"    - MAE Probability: {metrics['mae_prob']:.2e}")
            print(f"    - Avg batch time: {metrics['avg_time']:.4f}s")
            print()

        except Exception as e:
            print(f"  ✗ Failed: {e}")
            print()
            continue

    # Print summary table
    if results:
        print("=" * 80)
        print("SUMMARY - Training vs Serving Comparison")
        print("=" * 80)
        print()
        print("Comparing serving predictions against training pipeline predictions:")
        print("- Tolerance: 1e-6 (probabilities must match within 0.000001)")
        print(f"- Sample size: {len(truth_preds)} predictions")
        print()
        print(format_results_table(results, tolerance=1e-6))
        print()
        print("Legend:")
        print("  Pred Match %: Percentage of predictions that exactly match training")
        print(
            "  Prob Match %: Percentage of samples where ALL probabilities are within 1e-6"
        )
        print("  Max Diff: Maximum absolute difference in any probability")
        print("  MAE: Mean Absolute Error across all probabilities")
        print()

        print("=" * 80)
        print("SKEW ANALYSIS")
        print("=" * 80)
        print()

        perfect_services = [
            name
            for name, metrics in results.items()
            if metrics["prediction_match_rate"] == 1.0
            and metrics["prob_within_tolerance"]
        ]

        if perfect_services:
            print(f"ZERO SKEW (< 1e-6): {', '.join(perfect_services)}")
            print(
                "  All predictions and probabilities match training exactly within tolerance."
            )

        near_perfect = [
            name
            for name, metrics in results.items()
            if metrics["prediction_match_rate"] == 1.0
            and not metrics["prob_within_tolerance"]
            and metrics["max_prob_diff"] < 1e-3
        ]

        if near_perfect:
            print(f"\nMinimal skew (< 1e-3): {', '.join(near_perfect)}")
            print(
                "  Predictions match exactly, probabilities have minor floating-point differences."
            )
            for name in near_perfect:
                m = results[name]
                print(
                    f"    {name}: max diff = {m['max_prob_diff']:.2e}, "
                    f"{m['prob_match_rate'] * 100:.1f}% samples within 1e-6"
                )

        significant_skew = [
            name
            for name, metrics in results.items()
            if metrics["prediction_match_rate"] < 0.99
            or metrics["max_prob_diff"] > 0.01
        ]

        if significant_skew:
            print(f"\nSIGNIFICANT SKEW DETECTED: {', '.join(significant_skew)}")
            print("  This indicates issues with preprocessing or model conversion!")
            for name in significant_skew:
                m = results[name]
                print(f"    {name}:")
                print(
                    f"      - Prediction match: {m['prediction_match_rate'] * 100:.2f}%"
                )
                print(f"      - Prediction mismatches: {m['prediction_mismatches']}")
                print(f"      - Max probability diff: {m['max_prob_diff']:.2e}")

        print()

        print("=" * 80)
        print("PERFORMANCE COMPARISON")
        print("=" * 80)
        print()
        if len(results) > 1:
            times = {name: metrics["avg_time"] for name, metrics in results.items()}
            fastest = min(times, key=lambda x: times[x])
            slowest = max(times, key=lambda x: times[x])
            speedup = times[slowest] / times[fastest]

            print(f"Fastest: {fastest} ({times[fastest]:.4f}s per batch)")
            print(f"Slowest: {slowest} ({times[slowest]:.4f}s per batch)")
            print(f"Speedup: {speedup:.2f}x")
            print()

            if "FastAPI + Pickle" in times:
                baseline = times["FastAPI + Pickle"]
                print("Speedup vs Pickle baseline:")
                for name, time in sorted(times.items(), key=lambda x: x[1]):
                    if name != "FastAPI + Pickle":
                        improvement = baseline / time
                        print(f"  {name}: {improvement:.2f}x faster")
            print()


if __name__ == "__main__":
    main()
