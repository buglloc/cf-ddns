FROM golang:1.21.5 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/cf-ddns ./cmd/cf-ddns

FROM debian:bookworm-slim

COPY --from=build /go/bin/cf-ddns /

ENTRYPOINT ["/cf-ddns"]
