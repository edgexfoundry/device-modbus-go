# Device Modbus Go
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
