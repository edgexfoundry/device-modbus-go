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

// ==================== Tests for splitGroupReadResponse ====================

func TestSplitGroupReadResponse_SingleRequest(t *testing.T) {
	// Setup: single request for UINT16 in INPUT_REGISTERS
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Simulate data: 2 bytes for one UINT16 register
	data := []byte{0x00, 0x42} // Value: 66

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 0, results[0].OriginalIdx)
	assert.Equal(t, []byte{0x00, 0x42}, results[0].Data)
}

func TestSplitGroupReadResponse_TwoContiguousRequests(t *testing.T) {
	// Setup: two contiguous UINT16 requests
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
		mockCommandRequest(INPUT_REGISTERS, UINT16, 11),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Simulate data: 4 bytes for two UINT16 registers
	data := []byte{0x00, 0x0A, 0x00, 0x14} // Values: 10, 20

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, 0, results[0].OriginalIdx)
	assert.Equal(t, []byte{0x00, 0x0A}, results[0].Data)

	assert.Equal(t, 1, results[1].OriginalIdx)
	assert.Equal(t, []byte{0x00, 0x14}, results[1].Data)
}

func TestSplitGroupReadResponse_MixedRegisterSizes(t *testing.T) {
	// Setup: UINT16 followed by UINT32 (contiguous)
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 0),
		mockCommandRequestWithType(HOLDING_REGISTERS, INT32, 1),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	// These won't be grouped together due to different data types
	assert.Len(t, groups, 2)

	// Test first group (UINT16)
	data1 := []byte{0x01, 0x02}
	results1, err := splitGroupReadResponse(data1, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results1, 1)
	assert.Equal(t, []byte{0x01, 0x02}, results1[0].Data)

	// Test second group (UINT32 - 4 bytes)
	data2 := []byte{0x03, 0x04, 0x05, 0x06}
	results2, err := splitGroupReadResponse(data2, groups[1], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results2, 1)
	assert.Equal(t, []byte{0x03, 0x04, 0x05, 0x06}, results2[0].Data)
}

func TestSplitGroupReadResponse_MultipleContiguousINT32(t *testing.T) {
	// Setup: three contiguous INT32 requests
	reqs := []sdkModel.CommandRequest{
		mockCommandRequestWithType(HOLDING_REGISTERS, INT32, 0),
		mockCommandRequestWithType(HOLDING_REGISTERS, INT32, 2),
		mockCommandRequestWithType(HOLDING_REGISTERS, INT32, 4),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Simulate data: 12 bytes for three UINT32 registers (each uses 2 registers = 4 bytes)
	data := []byte{
		0x00, 0x00, 0x00, 0x01, // 1
		0x00, 0x00, 0x00, 0x02, // 2
		0x00, 0x00, 0x00, 0x03, // 3
	}

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x01}, results[0].Data)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x02}, results[1].Data)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x03}, results[2].Data)
}

func TestSplitGroupReadResponse_StringRegisters(t *testing.T) {
	// Setup: two contiguous string requests
	reqs := []sdkModel.CommandRequest{
		mockStringCommandRequest(INPUT_REGISTERS, 20, 2), // 2 registers = 4 bytes
		mockStringCommandRequest(INPUT_REGISTERS, 22, 3), // 3 registers = 6 bytes
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Simulate data: 10 bytes total (4 + 6)
	data := []byte{
		'A', 'B', 'C', 'D', // First string: "ABCD"
		'H', 'e', 'l', 'l', 'o', 0x00, // Second string: "Hello"
	}

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, []byte{'A', 'B', 'C', 'D'}, results[0].Data)
	assert.Equal(t, []byte{'H', 'e', 'l', 'l', 'o', 0x00}, results[1].Data)
}

func TestSplitGroupReadResponse_DataTooShort(t *testing.T) {
	// Setup: request expects 2 bytes
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Provide only 1 byte when 2 are needed
	data := []byte{0x00}

	_, err = splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data out of range")
}

func TestSplitGroupReadResponse_PreservesOriginalIndex(t *testing.T) {
	// Setup: requests in non-contiguous order that get reordered during aggregation
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 12), // Original index 0, will be second after sort
		mockCommandRequest(INPUT_REGISTERS, UINT16, 11), // Original index 1, will be first after sort
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// After sorting, order is: addr 11 (orig idx 1), addr 12 (orig idx 0)
	// Data for: addr 11, addr 12
	data := []byte{0x00, 0x0B, 0x00, 0x0C} // 11, 12

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// First result should be original index 1 (address 11)
	assert.Equal(t, 1, results[0].OriginalIdx)
	assert.Equal(t, []byte{0x00, 0x0B}, results[0].Data)

	// Second result should be original index 0 (address 12)
	assert.Equal(t, 0, results[1].OriginalIdx)
	assert.Equal(t, []byte{0x00, 0x0C}, results[1].Data)
}

func TestSplitGroupReadResponse_CoilsAndDiscreteInputs(t *testing.T) {
	// Setup: COILS (1 bit per register, packed into bytes)
	reqs := []sdkModel.CommandRequest{
		mockBoolCommandRequest(COILS, 0),
		mockBoolCommandRequest(COILS, 1),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	// Bool requests have length 1 each
	assert.Len(t, groups, 1)

	// For COILS/DISCRETE_INPUTS, byteLen = cmdInfo.Length (not * 2)
	// Each bool request has length 1, so byteLen = 1
	data := []byte{0x01, 0x00} // First coil on, second off

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, []byte{0x01}, results[0].Data)
	assert.Equal(t, []byte{0x00}, results[1].Data)
}

func TestSplitGroupReadResponse_EmptyData(t *testing.T) {
	// Setup: request expects data but none provided
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	data := []byte{}

	_, err = splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data out of range")
}

func TestSplitGroupReadResponse_ExactDataSize(t *testing.T) {
	// Setup: exact data size matches expected
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 0),
		mockCommandRequest(HOLDING_REGISTERS, UINT16, 1),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// Exactly 4 bytes for 2 UINT16 values
	data := []byte{0xFF, 0xFF, 0x00, 0x00}

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, []byte{0xFF, 0xFF}, results[0].Data)
	assert.Equal(t, []byte{0x00, 0x00}, results[1].Data)
}

func TestSplitGroupReadResponse_ExtraDataIgnored(t *testing.T) {
	// Setup: more data than needed (shouldn't cause error)
	reqs := []sdkModel.CommandRequest{
		mockCommandRequest(INPUT_REGISTERS, UINT16, 10),
	}
	reqsMeta, groups, err := aggregateReadRequests(reqs)
	assert.NoError(t, err)
	assert.Len(t, groups, 1)

	// 4 bytes when only 2 are needed
	data := []byte{0x00, 0x42, 0xFF, 0xFF}

	results, err := splitGroupReadResponse(data, groups[0], reqsMeta)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, []byte{0x00, 0x42}, results[0].Data)
}

// Helper function for creating CommandRequests with specific type
func mockCommandRequestWithType(primaryTable, rawType string, startingAddress uint16) sdkModel.CommandRequest {
	valueType := rawType
	// Map raw types to Go types for the Type field
	switch rawType {
	case UINT16:
		valueType = common.ValueTypeUint16
	case UINT32:
		valueType = common.ValueTypeUint32
	case UINT64:
		valueType = common.ValueTypeUint64
	case INT16:
		valueType = common.ValueTypeInt16
	case INT32:
		valueType = common.ValueTypeInt32
	case INT64:
		valueType = common.ValueTypeInt64
	case FLOAT32:
		valueType = common.ValueTypeFloat32
	case FLOAT64:
		valueType = common.ValueTypeFloat64
	}

	return sdkModel.CommandRequest{
		DeviceResourceName: fmt.Sprintf("TEST_%s_%s_%d", primaryTable, rawType, startingAddress),
		Type:               valueType,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    primaryTable,
			STARTING_ADDRESS: startingAddress,
			RAW_TYPE:         rawType,
		},
	}
}

// Helper function for creating Bool CommandRequests
func mockBoolCommandRequest(primaryTable string, startingAddress uint16) sdkModel.CommandRequest {
	return sdkModel.CommandRequest{
		DeviceResourceName: fmt.Sprintf("TEST_BOOL_%s_%d", primaryTable, startingAddress),
		Type:               common.ValueTypeBool,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    primaryTable,
			STARTING_ADDRESS: startingAddress,
		},
	}
}
