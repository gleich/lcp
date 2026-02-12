# syntax=docker/dockerfile:1
FROM golang:1.26.0 AS build

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/lcp ./cmd/lcp.go

FROM alpine:3.20.2

RUN apk update && apk add --no-cache ca-certificates=20250911-r0 tzdata=2025c-r0

WORKDIR /src
COPY --from=build /bin/lcp /bin/lcp

CMD ["/bin/lcp"]
