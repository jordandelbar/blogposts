FROM golang:1.24-alpine AS build-env
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags '-w -s' -a -o api ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=build-env /app/api .

EXPOSE 5000
CMD ["./api"]
