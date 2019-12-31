FROM golang:latest

RUN go get -u gitlab.com/stephane5/cloudflare-prometheus-exporter
CMD cloudflare-prometheus-exporter --api-email ${APIEMAIL} --api-key ${APIKEY}