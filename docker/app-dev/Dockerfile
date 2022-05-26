# syntax=docker/dockerfile:1

FROM golang:1.18.2

RUN go install github.com/cespare/reflex@v0.3.1
COPY reflex.conf /

COPY start.sh /

ENTRYPOINT ["reflex", "-c", "/reflex.conf"]