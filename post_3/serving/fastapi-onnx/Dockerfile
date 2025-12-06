FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir uv

COPY serving/fastapi-onnx/pyproject.toml ./

RUN uv pip install --system --no-cache -r pyproject.toml

COPY serving/fastapi-onnx/app/ ./app/
COPY serving/fastapi-onnx/main.py ./

RUN mkdir -p ./models

EXPOSE 8001

CMD ["gunicorn", "main:app", "--workers", "4", "--worker-class", "uvicorn.workers.UvicornWorker", "--bind", "0.0.0.0:8001", "--timeout", "120"]
