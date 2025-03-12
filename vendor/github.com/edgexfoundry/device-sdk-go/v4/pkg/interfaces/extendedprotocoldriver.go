// Copyright (C) 2024 IOTech Ltd

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// ExtendedProtocolDriver is a low-level device-specific interface implemented
// by device services that support extended ProtocolDriver features.
type ExtendedProtocolDriver interface {
	// ProfileScan triggers specific device to discover device profile.
	ProfileScan(req requests.ProfileScanRequest) (model.DeviceProfile, error)
	// StopDeviceDiscovery stops the ongoing device discovery process.
	StopDeviceDiscovery(options map[string]any)
	// StopProfileScan stops the ongoing device profile scan process for a specific device.
	StopProfileScan(deviceName string, options map[string]any)
}
