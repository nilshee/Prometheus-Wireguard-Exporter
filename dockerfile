FROM golang:1.25 AS builder

WORKDIR /app

COPY /src .

RUN go mod vendor
RUN go build -ldflags "-linkmode external -extldflags '-static'" -o main  -mod=vendor ./cmd


FROM ubuntu:24.04

COPY ./wrieguard_setup.sh .

RUN chmod +x ./wrieguard_setup.sh && \
    ./wrieguard_setup.sh

WORKDIR /opt/

COPY --from=builder /app/main .
