# Cloudflare prometheus exporter

Cloudflare prometheus exporter provides direct support for Cloudflare metrics to be exposed to Prometheus.

This library uses Cloudflare's GraphQL endpoint to fetch the aggregated metrics for a given zone/user(future).

### Prometheus metrics

## Supported metrics

- Caching (cached, uncached)

## Format

Here are the exposed metrics

- **cloudflare_done_fetches**: Number of fetchs effectively successfully sent towards the API 
- **cloudflare_failed_fetches**: Number of fetchs having failed
- **cloudflare_processed_bytes{*cacheStatus*}**: Number of bytes per *cacheStatus*

Here is a sample of metric you should get once running and fetching from the API

```
# HELP cloudflare_done_fetches The total number of done fetches
# TYPE cloudflare_done_fetches counter
cloudflare_done_fetches 1
# HELP cloudflare_failed_fetches The total number of failed fetches
# TYPE cloudflare_failed_fetches counter
cloudflare_failed_fetches 0
# HELP cloudflare_processed_bytes The total number of processed bytes, labelled per cache status
# TYPE cloudflare_processed_bytes gauge
cloudflare_processed_bytes{cacheStatus="dynamic"} 7662
cloudflare_processed_bytes{cacheStatus="hit"} 72128
```

Cache metrics are indexed with the cacheStatus as a label, so you can group by cacheStatus in your visualizations.

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

The exporter will record the metrics every 4 minutes for the interval from `time.Now() - 4 minutes -> time.Now()`

TODO:

- Argument to set frequency
- Auto-adapt frequency based on the volume of aggregated metrics

## Installation

```
go get -u gitlab.com/stephane5/cloudflare-prometheus-exporter
```

Once installed, call it as you would call any other GO binary 

```
cloudflare-prometheus-exporter <options>
```

## Docker machine

The Docker machine is publicly available on docker.io's registry at this address: https://hub.docker.com/repository/docker/stephanecloudflare/cloudflare-prometheus-exporter

```
docker run stephanecloudflare/cloudflare-prometheus-exporter -p 2112:2112 -e APIKEY=YOUR-KEY
5092dbe60 -e APIEMAIL=YOUR-EMAIL
```

**Note**: the exposed port could be the one you wish to use externally but the service itself should be kept on 2112 TCP (default port hard coded in the script)

## Support

The project is a personal project and hence Cloudflare support isn't going to be able to provide support for it, please submit your requests directly toward the issue section of this repository.

**Note**: Cloudflare responsability will not be engaged for any issues you may encounter using this open-source project, you use it at your own risks and downloading this project and execute it worth agreement of this statement by the user.

More metrics are going to arrive while I'm stabilising the process, feel free to PR your changes directly against the repo for anything you want to add.
