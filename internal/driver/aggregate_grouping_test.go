package driver

import (
	"fmt"
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/stretchr/testify/assert"
)

func TestAggregateReadRequests_SingleGroup(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 11),
		mockCommandRequest(INPUT_REGISTERS, UINT16, 12),
	}
	rs, groups, err := aggregateReadRequests(reqs)
	assert.Len(t, rs, 2)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, 0, groups[0].startIdx)
	assert.Equal(t, 1, groups[0].endIdx)
	assert.Equal(t, uint16(11), groups[0].startAddr)
	assert.Equal(t, uint16(2), groups[0].totalLen)
}

func TestAggregateReadRequests_MultipleGroups(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 0),
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 2),
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
		mockCommandRequest(INPUT_REGISTERS, UINT16, 11),
	}
	rq, groups, err := aggregateReadRequests(reqs)
	assert.Len(t, rq, 4)
	assert.NoError(t, err)
	assert.Len(t, groups, 3)
	assert.Equal(t, 0, groups[0].startIdx)
	assert.Equal(t, 0, groups[0].endIdx)
	assert.Equal(t, uint16(0), groups[0].startAddr)
	assert.Equal(t, uint16(1), groups[0].totalLen)

	assert.Equal(t, 1, groups[1].startIdx)
	assert.Equal(t, 1, groups[1].endIdx)
	assert.Equal(t, uint16(2), groups[1].startAddr)
	assert.Equal(t, uint16(1), groups[1].totalLen)

	assert.Equal(t, 2, groups[2].startIdx)
	assert.Equal(t, 3, groups[2].endIdx)
	assert.Equal(t, uint16(10), groups[2].startAddr)
	assert.Equal(t, uint16(2), groups[2].totalLen)
}

func TestAggregateReadRequests_NonContiguous(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 0),
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 5), // not contiguous
	}
	_, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 2)
}

func TestAggregateReadRequests_DifferentTypes(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 0),
		mockCommandRequest(HOLDING_REGISTERS, INT16, 1), // different valueType
	}
	_, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 2)
}

func TestAggregateReadRequests_Empty(t *testing.T) {
	reqs := []sdkModel.CommandRequest{}
	_, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 0)
}

func TestAggregateReadRequests_StringValues_SingleGroup(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockStringCommandRequest(INPUT_REGISTERS, 20, 5),
		mockStringCommandRequest(INPUT_REGISTERS, 25, 5), // contiguous string registers
	}
	rs, groups, err := aggregateReadRequests(reqs)
	assert.Len(t, rs, 2)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, 0, groups[0].startIdx)
	assert.Equal(t, 1, groups[0].endIdx)
	assert.Equal(t, uint16(20), groups[0].startAddr)
	assert.Equal(t, uint16(10), groups[0].totalLen) // 5+5
}

func TestAggregateReadRequests_StringValues_MultipleGroups(t *testing.T) {
	reqs := []sdkModel.CommandRequest{
		mockStringCommandRequest(HOLDING_REGISTERS, 0, 4),
		mockStringCommandRequest(HOLDING_REGISTERS, 10, 4), // not contiguous
		mockStringCommandRequest(INPUT_REGISTERS, 20, 2),
	}
	rs, groups, err := aggregateReadRequests(reqs)
	assert.Len(t, rs, 3)
	assert.NoError(t, err)
	assert.Len(t, groups, 3)
	assert.Equal(t, 0, groups[0].startIdx)
	assert.Equal(t, 0, groups[0].endIdx)
	assert.Equal(t, uint16(0), groups[0].startAddr)
	assert.Equal(t, uint16(4), groups[0].totalLen)

	assert.Equal(t, 1, groups[1].startIdx)
	assert.Equal(t, 1, groups[1].endIdx)
	assert.Equal(t, uint16(10), groups[1].startAddr)
	assert.Equal(t, uint16(4), groups[1].totalLen)

	assert.Equal(t, 2, groups[2].startIdx)
	assert.Equal(t, 2, groups[2].endIdx)
	assert.Equal(t, uint16(20), groups[2].startAddr)
	assert.Equal(t, uint16(2), groups[2].totalLen)
}

// mockCommandRequest creates a CommandRequest with minimal fields for grouping
func mockCommandRequest(primaryTable, rawType string, startingAddress uint16) sdkModel.CommandRequest {
	return sdkModel.CommandRequest{
		DeviceResourceName: fmt.Sprintf("TEST_%s_%s_%d", primaryTable, rawType, startingAddress),
		Type:               rawType,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    primaryTable,
			STARTING_ADDRESS: startingAddress,
			RAW_TYPE:         rawType,
		},
	}
}

// mockStringCommandRequest creates a CommandRequest for string values
func mockStringCommandRequest(primaryTable string, startingAddress, registerSize uint16) sdkModel.CommandRequest {
	return sdkModel.CommandRequest{
		DeviceResourceName: fmt.Sprintf("TEST_STRING_%s_%d_%d", primaryTable, startingAddress, registerSize),
		Type:               common.ValueTypeString,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:        primaryTable,
			STARTING_ADDRESS:     startingAddress,
			RAW_TYPE:             UINT16,
			STRING_REGISTER_SIZE: registerSize,
		},
	}
}
