FROM debian:buster-slim

ARG TARGETPLATFORM

WORKDIR /app

COPY . .

RUN set -ex \
  && if [ "${TARGETPLATFORM}" = "linux/amd64" ]; then export TARGETPLATFORM=amd64; fi \
  && if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then export TARGETPLATFORM=arm64; fi \
  && mv ocean /usr/local/bin/tdex \
  && mv "oceand-linux-$TARGETPLATFORM" /usr/local/bin/tdexd


# $USER name, and data $DIR to be used in the `final` image
ARG USER=ocean
ARG DIR=/home/ocean

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

# NOTE: Default GID == UID == 1000
RUN adduser --disabled-password \
            --home "$DIR/" \
            --gecos "" \
            "$USER"
USER $USER

# Prevents `VOLUME $DIR/.oceand/` being created as owned by `root`
RUN mkdir -p "$DIR/.oceand/"

# Expose volume containing all `oceand` data
VOLUME $DIR/.oceand/

# expose trader and operator interface ports
EXPOSE 18000
EXPOSE 18001

ENTRYPOINT [ "oceand" ]
