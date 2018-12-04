#
# Copyright (C) 2018 IOTech Ltd
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.11.2-alpine3.7 AS builder

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

RUN echo http://nl.alpinelinux.org/alpine/v3.6/main > /etc/apk/repositories; \
    echo http://nl.alpinelinux.org/alpine/v3.6/community >> /etc/apk/repositories

RUN apk update && apk add make && apk add bash
RUN apk add curl && apk add git && apk add openssh && apk add build-base

# RUN curl https://glide.sh/get | sh
# Workaround for the requirement to pin the version of glide to v0.13.1
ADD getglide.sh .
RUN sh ./getglide.sh



# set the working directory
WORKDIR $GOPATH/src/github.com/edgexfoundry/device-modbus-go

COPY . .

RUN make prepare
RUN make build


FROM scratch

ENV APP_PORT=49991
EXPOSE $APP_PORT

COPY --from=builder /go/src/github.com/edgexfoundry/device-modbus-go/cmd /

ENTRYPOINT ["/device-modbus","--registry","--profile=docker","--confdir=/res"]
