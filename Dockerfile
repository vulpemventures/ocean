# first image used to build the sources
FROM golang:1.20-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.Version=${COMMIT}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${COMMIT}'" -o bin/oceand cmd/oceand/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/ocean cmd/ocean/*

# Second image, running the oceand executable
FROM debian:buster-slim

# Set the working directory inside the container
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /app/bin/* /app
COPY --from=builder /app/internal/infrastructure/storage/db/postgres/migration/* /app

ENV OCEAN_DB_MIGRATION_PATH=file://
ENV OCEAN_DATADIR=/app/data/oceand 
ENV OCEAN_CLI_DATADIR=/app/data/ocean
ENV PATH="/app:${PATH}"

# Expose volume containing all `oceand` data
VOLUME /app/data/oceand
VOLUME /app/data/ocean

# Expose ports of grpc server and profiler
EXPOSE 18000
EXPOSE 18001

ENTRYPOINT ["oceand"]

