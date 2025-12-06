.PHONY: training run-fa-pickle run-fa-onnx run-axum-onnx load-testing-stress \
	load-testing-normal build up down

USERS ?= 100
RATE  ?= 100
TIME  ?= 30s
HOST  ?= 8000
NTRIALS ?= 100
PERF_FREQ ?= 99

training:
	cd training && uv run main.py --n-trials $(NTRIALS)

run-fa-pickle:
	cd serving/fastapi-pickle && uv run uvicorn main:app --workers 4 --log-level warning --no-access-log

run-fa-onnx:
	cd serving/fastapi-onnx && uv run uvicorn main:app --workers 4 --log-level warning --no-access-log

run-axum-onnx:
	cd serving/axum-onnx && cargo build --release && ./target/release/ort-onnx

load-testing-stress:
	cd load-testing && uv run locust -f locustfile.py \
		--host http://localhost:$(HOST) \
		--headless \
		-u 500 \
		-r 100 \
		-t 30s

load-testing-normal:
	cd load-testing && uv run locust -f locustfile.py \
	    --host http://localhost:$(HOST) \
	    --headless \
		-u 50 \
	    -r 5 \
		-t 30s \
	    MixedUser

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down
