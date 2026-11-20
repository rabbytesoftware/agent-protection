# Agent Protection -- local-use image, no Go/Rust/Node toolchain required.
# CI uses the composite Action (binary download) instead; this image exists
# for contributors running `ap check .` locally without installing anything.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /ap ./cmd/ap

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /ap /usr/local/bin/ap
WORKDIR /workspace
ENTRYPOINT ["ap"]
CMD ["check", "."]
