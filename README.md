# Device Modbus Go
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-modbus-go/job/master/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-modbus-go/job/master/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-modbus-go/branch/master/graph/badge.svg?token=tgWsR3KWGX)](https://codecov.io/gh/edgexfoundry/device-modbus-go) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-modbus-go)](https://goreportcard.com/report/github.com/edgexfoundry/device-modbus-go) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-modbus-go?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-modbus-go/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-modbus-go?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-modbus-go)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-modbus-go) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-modbus-go)](https://github.com/edgexfoundry/device-modbus-go/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-modbus-go)](https://github.com/edgexfoundry/device-modbus-go/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-modbus-go-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-modbus-go)](https://github.com/edgexfoundry/device-modbus-go/commits)

## Overview
Modbus Micro Service - device service for connecting Modbus devices to EdgeX.
## Usage
Users can refer to [the document](https://docs.edgexfoundry.org/1.2/examples/Ch-ExamplesAddingModbusDevice) to learn how to use this device service.
## Example Profile and Device
The `ProfilesDir` and `DevicesDir` in the configuration.toml are empty string by default.
To use the example Profile and Device in this repository, please fill './res/profiles' and './res/devices'
to `ProfilesDir` and `DevicesDir` respectively.
`modbus.test.device.profile.toml` and `modbus.test.devices.toml` will be loaded and created when the Device Service starts up.
Users can modify those files or add additional Profile YAML or Device TOML to meet their needs.
## Modbus Simulator
Build and run the Modbus simulator
```
$ cd simulator
$ go build
$ ./simulator 
Modbus TCP address: 0.0.0.0:1502 
Start up a Modbus TCP simulator.
```

## Community
- Chat: https://edgexfoundry.slack.com
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
