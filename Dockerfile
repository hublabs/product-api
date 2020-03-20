FROM golang:1.13 AS builder

RUN go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /go/src/github.com/hublabs/product-api
ADD go.mod go.sum ./
RUN go mod download
ADD . /go/src/github.com/hublabs/product-api
ENV CGO_ENABLED=0
RUN go build -o product-api

FROM pangpanglabs/alpine-ssl
WORKDIR /go/src/github.com/hublabs/product-api
COPY --from=builder /go/src/github.com/hublabs/product-api/*.yml /go/src/github.com/hublabs/product-api/
COPY --from=builder /go/src/github.com/hublabs/product-api/product-api /go/src/github.com/hublabs/product-api/
COPY --from=builder /go/src/github.com/hublabs/product-api/run.sh /go/src/github.com/hublabs/product-api/
RUN chmod +x ./run.sh

EXPOSE 5000

CMD ["/bin/sh","./run.sh"]