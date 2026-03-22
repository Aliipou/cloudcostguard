FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT_SHA=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/Aliipou/cloudcostguard/cmd.Version=${VERSION} -X github.com/Aliipou/cloudcostguard/cmd.CommitSHA=${COMMIT_SHA} -X github.com/Aliipou/cloudcostguard/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /cloudcostguard .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

COPY --from=builder /cloudcostguard /usr/local/bin/cloudcostguard

USER appuser

ENTRYPOINT ["cloudcostguard"]
