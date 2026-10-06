# syntax=docker/dockerfile:1

FROM build-env

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
