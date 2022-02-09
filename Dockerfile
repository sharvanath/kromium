# syntax = docker/dockerfile:1.2

FROM golang:1.16.13-alpine AS build
WORKDIR /src
RUN apk add --no-cache file
ENV GOMODCACHE /root/.cache/gocache
RUN --mount=target=. --mount=target=/root/.cache,type=cache \
    CGO_ENABLED=0 go build -o /out/kromium -ldflags '-s -d -w'; \
    file /out/kromium | grep "statically linked"

FROM scratch
COPY --from=build /out/kromium /bin/kromium
ENTRYPOINT ["/bin/kromium"]
