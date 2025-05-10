FROM golang:1.24.3-alpine3.21 AS builder

ARG EXEC_FILE=chatbot-go

WORKDIR /data/app/

COPY . .
RUN sh build.sh

FROM alpine:3.21

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /data/app/
COPY --from=builder /data/app/$EXEC_FILE /data/app/configs/* ./

ENV COMMONCONF=common.yaml APPCONF=app.yaml PORT=80

CMD ["sh", "-c", "./$EXEC_FILE -common.conf $COMMONCONF -app.conf $APPCONF -http.port $PORT"]
