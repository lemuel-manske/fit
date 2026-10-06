# syntax=docker/dockerfile:1

FROM workspace AS build

ARG GOOS=linux
ARG GOARCH=amd64
ARG BINARY_NAME=fit

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -o "/out/$BINARY_NAME" .

FROM scratch AS artifact

COPY --from=build /out/ /
