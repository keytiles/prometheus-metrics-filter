# ToDo  own images, with curl for healthcheck...

ARG LABEL_VERSION=1.0.0

# build stage

FROM golang:1.21.1-bookworm AS builder

ARG LABEL_VERSION
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN test -n "${LABEL_VERSION}"

WORKDIR /app
COPY . .

# attilaw: removed mod=vendor as caused issues (investigate later)
#RUN go build -v -mod=vendor -ldflags "-s -w -X main.version=${LABEL_VERSION}" -o prometheus-metrics-filter
RUN go build -v -ldflags "-s -w -X main.version=${LABEL_VERSION}" -o prometheus-metrics-filter

# final stage

FROM alpine:3.18
RUN apk --no-cache add ca-certificates make
# RUN apt-get update \
#     && apt-get install -y -q --no-install-recommends curl  \
#     && rm -r /var/lib/apt/lists/*

# copy over the compiled binary
COPY --from=builder /app/prometheus-metrics-filter /app/prometheus-metrics-filter
# copy over the config files
COPY --from=builder /app/docker/files/scylladb-config.yaml /conf/config.yaml
COPY --from=builder /app/docker/files/log-config.yaml /conf/log-config.yaml

# we start the app the way it takes the 2 (previously) copied config files
CMD ["/app/prometheus-metrics-filter"]