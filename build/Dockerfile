FROM golang:1.23-alpine as builder
RUN set -xe  && \
    apk update && apk upgrade  && \
    apk add --no-cache make git

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ARG APP

COPY . /go/src/app
WORKDIR /go/src/app

RUN if [[ "$APP" == "server" ]] ; then make build_server ; else make build_client ; fi

RUN cp bin/app /app

FROM alpine
RUN set -xe  && \
    apk update && apk upgrade  && \
    apk add --no-cache make git
COPY --from=builder /app /app

CMD ["/app"]