FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY internal ./internal
COPY cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/vector-engine ./cmd/vector-engine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/preprocess ./cmd/preprocess


FROM builder AS preprocess

COPY resources /app/resources

RUN mkdir -p /app/processed && /out/preprocess


FROM scratch AS api

COPY --from=builder /out/api /api
COPY resources/mcc_risk.json /app/resources/mcc_risk.json
COPY resources/normalization.json /app/resources/normalization.json

ENTRYPOINT ["/api"]


FROM scratch AS vector-engine

COPY --from=builder /out/vector-engine /vector-engine
COPY --from=preprocess /app/processed /app/processed

ENTRYPOINT ["/vector-engine"]
