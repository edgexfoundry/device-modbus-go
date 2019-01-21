package driver

import (
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestCheckAttributes(t *testing.T) {
	requests := []sdkModel.CommandRequest{
		{
			DeviceObject: models.DeviceObject{
				Attributes: map[string]interface{}{
					"primaryTable":    "HOLDING_REGISTERS",
					"startingAddress": 1001,
					"isByteSwap":      "true"},
			},
		},
		{
			DeviceObject: models.DeviceObject{
				Attributes: map[string]interface{}{
					"primaryTable":    "HOLDING_REGISTERS",
					"startingAddress": "1002",
					"isByteSwap":      "true"},
			},
		},
	}

	err := checkAttributes(requests)
	if err != nil {
		t.Fatalf("Test check attributes failed! Error: %v", err)
	}
}

func TestCheckAttributes_fail(t *testing.T) {
	requests := []sdkModel.CommandRequest{
		{
			DeviceObject: models.DeviceObject{
				Attributes: map[string]interface{}{
					"startingAddress": 1001,
					"isByteSwap":      "true"},
			},
		},
		{
			DeviceObject: models.DeviceObject{
				Attributes: map[string]interface{}{
					"startingAddress": "1002",
					"isByteSwap":      "true"},
			},
		},
	}

	err := checkAttributes(requests)
	if err == nil {
		t.Fatalf("Test should be failed!")
	}
}
