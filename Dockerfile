#
# Copyright (C) 2018 IOTech Ltd
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.11-alpine AS builder

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

RUN apk add --update make git openssh build-base

# set the working directory
WORKDIR $GOPATH/src/github.com/edgexfoundry/device-modbus-go

COPY . .

RUN make build

FROM scratch

ENV APP_PORT=49991
EXPOSE $APP_PORT

COPY --from=builder /go/src/github.com/edgexfoundry/device-modbus-go/cmd /

ENTRYPOINT ["/device-modbus","--registry","--profile=docker","--confdir=/res"]
