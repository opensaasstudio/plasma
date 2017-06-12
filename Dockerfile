FROM gliderlabs/alpine:3.4

RUN apk --no-cache add ca-certificates openssl && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub && \
    wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.25-r0/glibc-2.25-r0.apk && \
    apk --no-cache add glibc-2.25-r0.apk

CMD ["/plasma/bin/plasma"]

WORKDIR /plasma

COPY ./template /plasma/template
COPY ./bin /plasma/bin

EXPOSE 8080 50051
