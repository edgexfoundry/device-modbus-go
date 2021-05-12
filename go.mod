module github.com/edgexfoundry/device-modbus-go

require (
	github.com/edgexfoundry/device-sdk-go/v2 v2.0.0-dev.58
	github.com/edgexfoundry/go-mod-core-contracts/v2 v2.0.0-dev.90
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/edgexfoundry/device-sdk-go/v2 => ../../device-sdk-go
)

go 1.16
