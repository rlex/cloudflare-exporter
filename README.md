# Cloudflare prometheus exporter

Cloudflare prometheus exporter provide direct support for Cloudflare metrics to be published in Prometheus.

This library uses Cloudflare's GraphQL endpoint to fetch the aggregated metrics for a given zone/user (future).

### Supported metrics

- Caching (cached, uncached)

## Usage

```
NAME:
   cloudflare-exporter - export Cloudflare metrics to prometheus

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --api-key value    Your Cloudflare API token
   --api-email value  The email address associated with your Cloudflare API token and account
   --help, -h         show help
   --version, -v      print the version
```

Once launched with valid credentials, the binary will spin a webserver on http://localhost:2112/metrics exposing the metrics received from Cloudflare's GraphQL endpoint.

The interval of check is set to 1 minute, during which the service will stay stale.

## Installation

```
go get -u gitlab.com/stephane5/cloudflare-prometheus-exporter
```

## Docker machine

```
docker run stephanecloudflare/cloudflare-prometheus-exporter -p 2112:2112 -e APIKEY=YOUR-KEY
5092dbe60 -e APIEMAIL=YOUR-EMAIL
```

Note: the exposed port could be the one you wish to use externally but the service itself should be kept on 2112 TCP (default port hard coded in the script)