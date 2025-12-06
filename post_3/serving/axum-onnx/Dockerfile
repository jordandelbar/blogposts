FROM rustlang/rust:nightly-bookworm-slim AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y \
    pkg-config \
    libssl-dev \
    g++ \
    libstdc++6 && \
    rm -rf /var/lib/apt/lists/*

COPY serving/ort-onnx/Cargo.toml serving/ort-onnx/Cargo.lock* ./

RUN mkdir -p src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release || true && \
    rm -rf src

COPY serving/ort-onnx/src ./src

RUN cargo build --release

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/ort-onnx /app/ort-onnx

RUN mkdir -p /app/models

ENV MODEL_PATH=/app/models/adult_income_model.onnx

EXPOSE 3000

CMD ["/app/ort-onnx"]
