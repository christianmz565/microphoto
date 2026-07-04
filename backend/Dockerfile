FROM golang:1.26-alpine AS deps

WORKDIR /app

RUN apk add --no-cache git vips-dev gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS builder-coordinator
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/coordinator ./cmd/coordinator

FROM deps AS builder-reaper
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/reaper ./cmd/reaper

FROM deps AS builder-worker
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/worker ./cmd/worker

FROM alpine:3.19 AS base
RUN apk add --no-cache ca-certificates vips

FROM base AS coordinator
RUN apk add --no-cache ffmpeg
COPY --from=builder-coordinator /bin/coordinator /usr/local/bin/coordinator
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/coordinator"]

FROM base AS reaper
COPY --from=builder-reaper /bin/reaper /usr/local/bin/reaper
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/reaper"]

FROM base AS worker
RUN apk add --no-cache ffmpeg
COPY --from=builder-worker /bin/worker /usr/local/bin/worker
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/worker"]
