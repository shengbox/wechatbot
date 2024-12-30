FROM golang:1.20 AS builder
ENV GOPROXY=https://proxy.golang.com.cn,direct
WORKDIR /app
ADD .. /app

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o wechatbot

FROM scratch AS final
LABEL maintainer="shengbox@gmail.com"

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Shanghai
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /go
COPY --from=builder /app/wechatbot .
COPY --from=builder /app/.env-sample .env
COPY --from=builder /app/functions.json .

ENTRYPOINT ["./wechatbot"]