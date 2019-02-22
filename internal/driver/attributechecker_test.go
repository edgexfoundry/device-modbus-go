package driver

import (
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestCheckAttributes(t *testing.T) {
	requests := []sdkModel.CommandRequest{
		{
			DeviceResource: models.DeviceResource{
				Attributes: map[string]interface{}{
					"primaryTable":    "HOLDING_REGISTERS",
					"startingAddress": 1001,
					"isByteSwap":      "true"},
			},
		},
		{
			DeviceResource: models.DeviceResource{
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
			DeviceResource: models.DeviceResource{
				Attributes: map[string]interface{}{
					"startingAddress": 1001,
					"isByteSwap":      "true"},
			},
		},
		{
			DeviceResource: models.DeviceResource{
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
