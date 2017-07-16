FROM openfresh/golang:1.8.3 AS build

WORKDIR /go/src/github.com/openfresh/plasma
COPY . . 
RUN make deps
RUN make build

FROM gliderlabs/alpine:3.6

WORKDIR /plasma
RUN apk --no-cache add ca-certificates openssl
COPY --from=build /go/src/github.com/openfresh/plasma/bin/plasma /plasma/bin/ 
COPY --from=build /go/src/github.com/openfresh/plasma/template/ /plasma/template/
CMD ["/plasma/bin/plasma"]
EXPOSE 8080 50051 9999
