plasma
==========

[![Circle CI](https://circleci.com/gh/openfresh/plasma.svg?style=shield&circle-token=8e3162c9a55b29a5fa0488834225a84c013977b1)](https://circleci.com/gh/openfresh/plasma)
[![Language](https://img.shields.io/badge/language-go-brightgreen.svg?style=flat)](https://golang.org/)
[![issues](https://img.shields.io/github/issues/openfresh/plasma.svg?style=flat)](https://github.com/openfresh/plasma/issues?state=open)
[![License: MIT](https://img.shields.io/badge/license-MIT-orange.svg)](LICENSE)
[![imagelayers.io](https://badge.imagelayers.io/openfresh/plasma:latest.svg)](https://imagelayers.io/?images=openfresh/plasma:latest 'Get your own badge on imagelayers.io')

![logo](img/plasma_logo.png)

plasma is event push middleware by using gRPC stream.

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
