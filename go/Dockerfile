FROM golang:1.19

WORKDIR /go

COPY . /go

RUN go env -w GO111MODULE=off

CMD ["/go/server"]