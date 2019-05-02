# Device Modbus Go
Modbus device service go version

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
  Name = "Modbus TCP test device"
  Profile = "Test.Device.Modbus.Profile"
  Description = "This device is a product for monitoring and controlling digital inputs and outputs over a LAN."
  labels = [ "Air conditioner","modbus TCP" ]
  [DeviceList.Protocols]
    [DeviceList.Protocols.modbus-tcp]
       Address = "0.0.0.0"
       Port = "1502"
       UnitID = "1"
  [[DeviceList.AutoEvents]]
    Frequency = "20s"
    OnChange = false
    Resource = "HVACValues"
```

**Modbus RTU**
```toml
[[DeviceList]]
  Name = "Modbus RTU test device"
  Profile = "Test.Device.Modbus.Profile"
  Description = "This device is a product for monitoring and controlling digital inputs and outputs over a LAN."
  labels = [ "Air conditioner","modbus RTU" ]
  [DeviceList.Protocols]
    [DeviceList.Protocols.modbus-rtu]
       Address = "/tmp/slave"
       BaudRate = "19200"
       DataBits = "8"
       StopBits = "1"
       Parity = "N"
       UnitID = "1"
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