#
# Copyright (c) 2020 IOTech Ltd
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

FROM golang:1.13-alpine AS builder

WORKDIR /go/src/github.com/edgexfoundry/device-modbus-go/simulator

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add zeromq-dev libsodium-dev pkgconfig build-base git

RUN git clone https://github.com/edgexfoundry/device-modbus-go.git

RUN mv ./device-modbus-go/simulator/* .

RUN CGO_ENABLED=0 GO111MODULE=on go build

FROM scratch

ENV APP_PORT=1502
EXPOSE $APP_PORT

COPY --from=builder /go/src/github.com/edgexfoundry/device-modbus-go/simulator /

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2020: IoTech Ltd'

ENTRYPOINT ["/simulator"]
