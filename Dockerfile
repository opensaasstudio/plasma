FROM gliderlabs/alpine:3.4

RUN apk --no-cache add ca-certificates openssl && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub && \
    apk --no-cache -X http://apkproxy.heroku.com/sgerrand/alpine-pkg-glibc add glibc glibc-bin

CMD ["/plasma/bin/plasma"]

WORKDIR /plasma

COPY ./template /plasma/template
COPY ./bin /plasma/bin

EXPOSE 8080
