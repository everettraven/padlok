FROM golang:1.26

COPY . .

RUN go build -o padlok main.go

ENTRYPOINT [ "./padlok" ]

