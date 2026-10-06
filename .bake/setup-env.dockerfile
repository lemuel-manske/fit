# syntax=docker/dockerfile:1

FROM golang:1.25-alpine

ENV CGO_ENABLED=0

WORKDIR /src
