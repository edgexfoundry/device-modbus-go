#
# Copyright (c) 2020-2021 IOTech Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG BASE=golang:1.15-alpine3.12
FROM ${BASE} AS builder

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache make git openssh gcc libc-dev zeromq-dev libsodium-dev

# set the working directory
WORKDIR /device-modbus-go

COPY . .

RUN go mod download

RUN make build

FROM alpine:3.12

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019-2021: IoTech Ltd'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache zeromq dumb-init

COPY --from=builder /device-modbus-go/cmd /
COPY --from=builder /device-modbus-go/LICENSE /
COPY --from=builder /device-modbus-go/Attribution.txt /

EXPOSE 49991

ENTRYPOINT ["/device-modbus"]
CMD ["--cp=consul://edgex-core-consul:8500", "--registry", "--confdir=/res"]
