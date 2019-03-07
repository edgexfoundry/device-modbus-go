package driver

import (
	"strings"
	"testing"

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func init() {
	driver = new(Driver)
	driver.Logger = logger.NewClient("test", false, "", "DEBUG")
}

func TestLockAddressWithAddressCountLimit(t *testing.T) {
	address := "/dev/USB0tty,19200,8,1,0"
	addressable := models.Addressable{
		Address: address,
	}
	driver.addressMap = make(map[string]chan bool)
	driver.workingAddressCount = make(map[string]int)
	driver.workingAddressCount[address] = concurrentCommandLimit

	err := driver.lockAddress(&addressable)

	if err == nil || !strings.Contains(err.Error(), "High-frequency command execution") {
		t.Errorf("Unexpect result, it should return high-frequency error, %v", err)
	}
}

func TestLockAddressWithAddressCountUnderLimit(t *testing.T) {
	address := "/dev/USB0tty,19200,8,1,0"
	addressable := models.Addressable{
		Address: address,
	}
	driver.addressMap = make(map[string]chan bool)
	driver.workingAddressCount = make(map[string]int)
	driver.workingAddressCount[address] = concurrentCommandLimit - 1

	err := driver.lockAddress(&addressable)

	if err != nil {
		t.Errorf("Unexpect result, address should be lock successfully, %v", err)
	}
}
