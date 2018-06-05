ARG ALPINE_VERSION=3.7
ARG GO_VERSION=1.10.2
ARG COMMIT=unknown
ARG TAG=unknown
ARG BUILDTIME=unknown

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build
RUN apk add --no-cache \
  bash \
  build-base \
  docker \
  git \
  util-linux
WORKDIR /go/src/github.com/docker/app/
COPY . .

FROM build AS bin-build
ARG COMMIT
ARG TAG
ARG BUILDTIME=unknown
RUN make COMMIT=${COMMIT} TAG=${TAG} BUILDTIME=${BUILDTIME} bin-all e2e-all

FROM build AS test
ARG COMMIT
ARG TAG
ARG BUILDTIME=unknown
RUN make COMMIT=${COMMIT} TAG=${TAG} BUILDTIME=${BUILDTIME} unit-test
