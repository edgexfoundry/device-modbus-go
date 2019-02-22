package driver

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

func checkAttributes(reqs []sdkModel.CommandRequest) error {
	var err error = nil
	for _, req := range reqs {
		attributes := req.DeviceResource.Attributes

		_, err = toString(attributes["primaryTable"])
		if err != nil {
			return fmt.Errorf("primaryTable fail to convert interface inoto string: %v", err)
		}

		_, err = toUint16(attributes["startingAddress"])
		if err != nil {
			return fmt.Errorf("startingAddress fail to convert interface inoto unit16: %v", err)
		}

		_, ok := attributes["isByteSwap"]
		if ok {
			_, err = toBool(attributes["isByteSwap"])
			if err != nil {
				return fmt.Errorf("isByteSwap fail to convert interface inoto boolean: %v", err)
			}
		}

		_, ok = attributes["isWordSwap"]
		if ok {
			_, err = toBool(attributes["isWordSwap"])
			if err != nil {
				return fmt.Errorf("isWordSwap fail to convert interface inoto boolean: %v", err)
			}
		}
	}
	return err
}
