#
# Copyright (c) 2020-2024 IOTech Ltd
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

ARG BASE=golang:1.23-alpine3.20
FROM ${BASE} AS builder

ARG ADD_BUILD_TAGS=""
ARG MAKE="make -e ADD_BUILD_TAGS=$ADD_BUILD_TAGS build"

RUN apk add --update --no-cache make git openssh

# set the working directory
WORKDIR /device-modbus-go

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN ${MAKE}

FROM alpine:3.20

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019-2024: IoTech Ltd'

RUN apk add --update --no-cache dumb-init
# Ensure using latest versions of all installed packages to avoid any recent CVEs
RUN apk --no-cache upgrade

COPY --from=builder /device-modbus-go/cmd /
COPY --from=builder /device-modbus-go/LICENSE /
COPY --from=builder /device-modbus-go/Attribution.txt /

EXPOSE 59901

ENTRYPOINT ["/device-modbus"]
CMD ["-cp=keeper.http://edgex-core-keeper:59890", "--registry"]
