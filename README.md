# Device Modbus Go
MQTT device service go version. The design is base on [ document](https://github.com/edgexfoundry/edgex-go/blob/master/docs/examples/Ch-ExamplesMQTTDeviceService.rst) .

## Requisite
* core-data
* core-metadata
* core-command

## Predefined configuration

### Device list
Define devices info for device-sdk to auto upload device profile and create device instance. Please modify `configuration.toml` file which under `./cmd/res` folder

**Modbus TCP**
```toml
[[DeviceList]]
  Name = "Damocles device"
  Profile = "CoolMasterNet Connected Device"
  Description = "Damocles2 is a product for monitoring and controlling digital inputs and outputs over a LAN."
  labels = [ "power-meter","Modbus TCP" ]
  [DeviceList.Addressable]
    name = "HVAC-Gateway address"
    Protocol = "TCP"
    Address = "127.0.0.1"
    Port = 1502
    Path = "1"
```

**Modbus RTU**
```toml
[[DeviceList]]
  Name = "Damocles device"
  Profile = "CoolMasterNet Connected Device"
  Description = "Damocles2 is a product for monitoring and controlling digital inputs and outputs over a LAN."
  labels = [ "power-meter","Modbus RTU" ]
  [DeviceList.Addressable]
    name = "HVAC-Gateway address"
    Protocol = "RTU"
    Address = "/tmp/slave,19200,8,1,0"
    Path = "1"
```

## Installation and Execution
```bash
make prepare
make build
make run
```


## build image
```bash
docker build -t edgexfoundry/docker-device-modbus-go:0.1.0 .
```