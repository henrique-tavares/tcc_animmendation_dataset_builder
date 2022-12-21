FROM golang:1.19 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o build/dataset_builder cmd/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/build/dataset_builder .

CMD ["./dataset_builder"]
