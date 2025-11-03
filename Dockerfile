FROM golang:1.24-alpine as builder

WORKDIR /app

ENV CGO_ENABLED=0

ARG GITHUB_SHA=dev

COPY . .

RUN go mod tidy
RUN go build -v -ldflags "-w -s -extldflags '-static' -X 'github.com/wuhan005/NekoBox/internal/conf.BuildCommit=$GITHUB_SHA'" -o NekoBox ./cmd/

FROM alpine:latest

RUN apk update && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo "Asia/Shanghai" > /etc/timezone

WORKDIR /home/app

COPY --from=builder /app/NekoBox .

RUN chmod 777 /home/app/NekoBox

ENTRYPOINT ["./NekoBox", "web"]
EXPOSE 8080
