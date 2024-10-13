# Traefik Auditor

This traefik plugin encapsulates the request and response (as observed by the traefik router) and sends it to a remote server after the fact.

This can be useful when you need to view the full request / response, for example:

  - Auditing requests to an application
  - Debugging user issues based on the request / response
  - Capturing requests for securit auditing / compliance
  - Capturing requests for performance analysis
  - Capturing requests for data streaming

You can use this plugin (alongside some code that you write in an external application) to add data streaming capabilities to your system.

**Note: By default, all request headers are sent in the payload. This means that it can be inherently DANGEROUS because `Authorization` headers are also sent
and can present a SIGNIFICANT SECURITY RISK - make sure that if you're testing, or consuming these messages somehow, that you sanitise your data accordingly.
This plugin allows for you to specifiy headers to ignore forwarding if you do not need them.**

## Why Use This Plugin?

Traefik logs the basic HTTP details to stdout, such as:

```bash
82.24.69.43 - - [12/Oct/2024:21:14:54 +0000] "POST /endpoints HTTP/1.1" 200 0 "-" "-" 6033 "default-svc-app-4b64bbe6c63a86082006@kubernetescrd" "http://10.11.11.41:8000" 13ms
```

These are useful for getting headlines, but sometimes you need more information about the request and response.

Combined with request tracing tools, this plugin can be used to provide a more complete view of the lifecycle.

## Remote Server

You need to configure a remote server to send logs to.

```yaml
traefik.http.middlewares.middleware-name.plugin.logger.remoteServer: http://localhost:8080/my/endpoint
```

You can also configure the timeout for the remote server:

```yaml
traefik.http.middlewares.middleware-name.plugin.logger.timeout: 1s
```

## Ignoring Headers

You can ignore headers in the request and response body by configuring `ignoreHeaders` param:

```yaml
traefik.http.middlewares.middleware-name.plugin.logger.ignoreHeaders: Authorization
```

Or to ignore multiple headers:

```yaml
traefik.http.middlewares.middleware-name.plugin.logger.ignoreHeaders: Authorization,X-Forwarded-For
```

If your service doe not respond with a `200` status code, this plugin will log that the attempt to send the record failed in the Traefik logs.


## Payload

This forwarder sends the request and response body to a remote server in a JSON POST request:

```json
{
  "duration": 31, // in milliseconds
  "request": {
    "time": "2024-10-13T10:19:03.44497217Z",
    "method": "GET",
    "content_length": 64,
    "path": "/my/endpoint",
    "query": {
      "raw": "a=b&a=c&x=z",
      "parsed": {
        "a": [
          "b",
          "c"
        ],
        "x": [
          "z"
        ]
      }
    },
    "headers": {
      "Connection": [
        "close"
      ],
      "Content-Type": [
        "application/json; charset=utf-8"
      ]
    },
    "body": "{\"request\":\"payload\"}"
  },
  "response": {
    "status": 200
    "time": "2024-10-13T10:19:03.475983461Z",
    "body": "{\"response\":\"payload\"}",
    "headers": {
      "Connection": [
        "close"
      ],
      "Content-Type": [
        "application/vnd.api+json"
      ]
    }
  }
}
```
