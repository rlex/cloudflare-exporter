FROM golang:latest

RUN go get -u gitlab.com/stephane5/cloudflare-prometheus-exporter
RUN echo ${APIEMAIL}
RUN echo ${APIKEY}
CMD cloudflare-prometheus-exporter --api-email ${APIEMAIL} --api-key ${APIKEY}