FROM golang:1.26 as builder

WORKDIR /src
COPY . .

RUN go build -o padlok main.go

FROM registry.access.redhat.com/ubi9-minimal:latest

COPY --from=builder /src/padlok /usr/local/bin/padlok
CMD ["padlok", "run"]

