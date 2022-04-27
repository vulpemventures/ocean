# first image used to build the sources
FROM golang:1.17-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG TARGETOS
ARG TARGETARCH


WORKDIR /oceand

COPY . .
RUN go mod download

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.Version=${COMMIT}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${COMMIT}'" -o oceand-linux cmd/oceand/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o ocean cmd/ocean/*

WORKDIR /build

RUN cp /oceand/oceand-linux .
RUN cp /oceand/ocean .

# Second image, running the oceand executable
FROM debian:buster

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /build/oceand-linux /
COPY --from=builder /build/ocean /

RUN install /ocean /bin
# Prevents `VOLUME $HOME/.oceand/` being created as owned by `root`
RUN useradd -ms /bin/bash user
USER user
RUN mkdir -p "$HOME/.oceand/"

# Expose ports of grpc server and profiler
EXPOSE 18000
EXPOSE 18001

CMD /oceand-linux

