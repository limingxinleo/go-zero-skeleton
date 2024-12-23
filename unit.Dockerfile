FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV GOPROXY=https://goproxy.cn,direct \
    ROOT_PATH=/app \
    CONFIG_PATH=etc/unit-api.yaml

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/main .

CMD ["./main"]
