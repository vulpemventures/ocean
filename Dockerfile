# first image used to build the sources
FROM golang:1.18-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.Version=${COMMIT}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${COMMIT}'" -o bin/oceand cmd/oceand/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/ocean cmd/ocean/*

# Second image, running the oceand executable
FROM debian:buster-slim

# $USER name, and data $DIR to be used in the `final` image
ARG USER=ocean
ARG DIR=/home/ocean

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /app/bin/* /usr/local/bin/

# NOTE: Default GID == UID == 1000
RUN adduser --disabled-password \
            --home "$DIR/" \
            --gecos "" \
            "$USER"
USER $USER

# Prevents `VOLUME $DIR/.oceand/` being created as owned by `root`
RUN mkdir -p "$DIR/.oceand/"

# Expose volume containing all `tdexd` data
VOLUME $DIR/.oceand/

# Expose ports of grpc server and profiler
EXPOSE 18000
EXPOSE 18001

ENTRYPOINT [ "oceand" ]

