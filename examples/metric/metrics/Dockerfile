FROM golang:1.18-alpine

COPY . /src/examples/metric/metrics
COPY ./../../.. /src/
WORKDIR /src

RUN go install /src/examples/metric/metrics/main.go
CMD ["/go/bin/main"]
