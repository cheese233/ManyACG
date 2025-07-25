FROM golang:alpine AS builder
WORKDIR /app

RUN apk add --no-cache git bash build-base pkgconf

RUN apk add --update --no-cache \
    --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main \
    vips-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILT_AT
ARG GIT_COMMIT
ARG VERSION

RUN builtAt=${BUILT_AT:-$(date +'%F %T %z')} && \
    gitCommit=${GIT_COMMIT:-$(git log --pretty=format:"%h" -1)} && \
    version=${VERSION:-$(git describe --abbrev=0 --tags)} && \
    ldflags="\
    -w -s \
    -X 'github.com/krau/ManyACG/common.BuildTime=$builtAt' \
    -X 'github.com/krau/ManyACG/common.Commit=$gitCommit' \
    -X 'github.com/krau/ManyACG/common.Version=$version'\
    " && \
    go build -ldflags "$ldflags" -o manyacg

FROM alpine:latest
WORKDIR /opt/manyacg/

RUN apk add --no-cache bash ca-certificates libc6-compat libgcc libstdc++ && update-ca-certificates

RUN apk add --update --no-cache \
    --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main \
    vips

COPY --from=builder /app/manyacg .
EXPOSE 39080

ENV MALLOC_ARENA_MAX=2 
ENTRYPOINT ["./manyacg"]
