plasma
==========

[![Circle CI](https://circleci.com/gh/openfresh/plasma.svg?style=shield&circle-token=8e3162c9a55b29a5fa0488834225a84c013977b1)](https://circleci.com/gh/openfresh/plasma)
[![Language](https://img.shields.io/badge/language-go-brightgreen.svg?style=flat)](https://golang.org/)
[![issues](https://img.shields.io/github/issues/openfresh/plasma.svg?style=flat)](https://github.com/openfresh/plasma/issues?state=open)
[![License: MIT](https://img.shields.io/badge/license-MIT-orange.svg)](LICENSE)
[![imagelayers.io](https://badge.imagelayers.io/openfresh/plasma:latest.svg)](https://imagelayers.io/?images=openfresh/plasma:latest 'Get your own badge on imagelayers.io')

![logo](img/plasma_logo.png)

plasma is event push middleware by using gRPC stream.


## Description

Plasma is middleware for sending event specialized for a stream. Plasma provides EventSource and gRPC Stream from the same endpoint.

## Installation

This middleware requires Redis.

### From Source

```bash
$ git clone git://github.com/openfresh/plasma.git $GOPATH/src/github.com/openfresh/plasma
$ cd  $GOPATH/src/github.com/openfresh/plasma
$ make deps
$ make build
```

The binary is generated under the `bin/` directory.

### Using docker

You can also use the Docker image.

```bash
$ docker run -p 8080:8080 openfresh/plasma
```

### Using docker-compose

You can use docker-compose for easy use without preparing Redis.

```bash
$ git clone git://github.com/openfresh/plasma.git $GOPATH/src/github.com/openfresh/plasma
$ cd  $GOPATH/src/github.com/openfresh/plasma
$ docker-compose up -d
```

# Usage Subscriber

## Sever Sent Event

[Using server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)

### GET /events

You request events that you want to subscribe to this endpoint. You can specify multiple events separated by commas.
The query name can be set with the `EventQuery` environment variable．(default value of EventQuery is  `eventType` )．

Here is a simple example using [Yaffle / EventSource] (https://github.com/Yaffle/EventSource).

```javascript
    var source = new EventSource('//localhost:8080/events?eventType=program:1234:views,program:1234:poll,program:1234:annotation');
    
    source.addEventListener("open", function(e) {
        console.log("open");
    });
    
    source.addEventListener("error", function(e) {
        console.log("error");
    });
    
    source.addEventListener("message", function(e) {
        console.log("message event: ", e.data);
    });
```

The JSON schema of data returned from Plasma is as follows.

```javascript
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "id": "/events",
    "properties": {
        "data": {
            "id": "/events/data",
            "type": "string"
        },
        "meta": {
            "id": "/events/meta",
            "properties": {
                "type": {
                    "id": "/events/meta/type",
                    "type": "string"
                }
            },
            "type": "object"
        }
    },
    "type": "object"
}
```

If the `DEBUG` environment variable is enabled, you can access the next two endpoints.

### GET /

This sample subscribes to `program:1234:views,program:1234:poll,program:1234:annotation` events.
When an event matching the subscribing event is published, an event is displayed on the page.

### GET /debug

You can publish events to Redis from this endpoint.  You need to enter valid JSON in EventData form.

## gRPC Stream

You can subscribe to events using gRPC Stream.

The ProtocolBuffer file is [here](https://github.com/openfresh/plasma/blob/master/protobuf/stream.proto) .

The following is a simple Go sample.

```go
func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := proto.NewStreamServiceClient(conn)
	ctx := context.Background()

	req := proto.Request{
		Events: []*proto.EventType{
			eventType("program:1234:poll"),
			eventType("program:1234:views"),
		},
	}
	
	ss, err := client.Events(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}
	
	for {
		resp, err := ss.Recv()
		if err != nil {
			log.Println(err)
			continue
		}
		if resp == nil {
			log.Println("payload is nil")
			continue
		}
		fmt.Printf("Meta: %s\tData: %s\n", resp.EventType.Type, resp.Data)
	}
}
```

# Usage Publisher

You publish events to the channel that Plasma subscribes according to the following JSON Schema.

```javascript
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "id": "/events",
    "properties": {
        "data": {
            "id": "/events/data",
            "type": "string"
        },
        "meta": {
            "id": "/events/meta",
            "properties": {
                "type": {
                    "id": "/events/meta/type",
                    "type": "string"
                }
            },
            "type": "object"
        }
    },
    "type": "object"
}
```

[openfresh/plasma-go](https://github.com/openfresh/plasma-go) is a library that wraps publish an event to Redis.

### ELB HealthCheck

Plasma supports the AWS Elastic Load Balancing health check.

If an User-Agent starts with `ELB-HealthChecker/`, Plasma will check health.

## Config

| name                                            | type          | desc                                                                                  | default        | note                                                                               |
|-------------------------------------------------|---------------|---------------------------------------------------------------------------------------|----------------|------------------------------------------------------------------------------------|
| PLASMA_PORT                                     | string        | port number                                                                           | 8080           |                                                                                    |
| PLASMA_ORIGIN                                   | string        | set to Access-Controll-Allow-Origin                                                   |                |                                                                                    |
| PLASMA_SSE_RETRY                                | int           | reconnect to the source milliseconds after each connection is closed                  | 2000           |                                                                                    |
| PLASMA_SSE_EVENTQUERY                           | string        | use as a querystring in SSE                                                           | eventType      | ex) /?eventType=program:1234:views                                                 |
| PLASMA_SUBSCRIBER_TYPE                          | string        | subscriber type                                                                       | mock           | support "mock" and "redis"                                                         |
| PLASMA_SUBSCRIBER_REDIS_ADDR                    | string        | Redis address including port number                                                   | localhost:6379 |                                                                                    |
| PLASMA_SUBSCRIBER_REDIS_PASSWORD                | string        | Redis password                                                                        |                |                                                                                    |
| PLASMA_SUBSCRIBER_REDIS_DB                      | int           | Redis DB                                                                              | 0              |                                                                                    |
| PLASMA_SUBSCRIBER_REDIS_CHANNELS                | string        | channels of Redis to subscribe (multiple specifications possible)                     |                |                                                                                    |
| PLASMA_SUBSCRIBER_REDIS_OVER_MAX_RETRY_BEHAVIOR | string        | Behavior of plasma when the number of retries connecting to Redis exceeds the maximum |                | "die" or "alive"                                                                   |
| PLASMA_SUBSCRIBER_REDIS_TIMEOUT                 | time.Duration | timeout for receive message from Redis                                                | 1s             |                                                                                    |
| PLASMA_SUBSCRIBER_REDIS_RETRY_INTERVAL          | time.Duration | interval for retry to receive message from Redis                                      | 5s             |                                                                                    |
| PLASMA_ERROR_LOG_OUT                            | string        | log file path                                                                         |                | stdout, stderr, filepath                                                           |
| PLASMA_ERROR_LOG_LEVEL                          | string        | log output level                                                                      |                | panic,fatal,error,warn,info,debug                                                  |
| PLASMA_ACCESS_LOG_OUT                           | string        | log file path                                                                         |                | stdout, stderr, filepath                                                           |
| PLASMA_ACCESS_LOG_LEVEL                         | string        | log output level                                                                      |                | panic,fatal,error,warn,info,debug                                                  |
| PLASMA_TLS_CERT_FILE                            | string        | cert file path                                                                        |                | TLS is enabled only when you set both PLASMA_TLS_CERT_FILE and PLASMA_TLS_KEY_FILE |
| PLASMA_TLS_KEY_FILE                             | string        | key file path                                                                         |                |                                                                                    |


License
===
See [LICENSE](LICENSE).

Copyright © CyberAgent, Inc. All Rights Reserved.
