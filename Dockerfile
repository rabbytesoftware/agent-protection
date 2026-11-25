FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /ap ./cmd/ap

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /ap /usr/local/bin/ap
WORKDIR /workspace
ENTRYPOINT ["ap"]
CMD ["check", "."]
