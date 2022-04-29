package cosmwasm

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/stretchr/testify/require"
)

func TestGetPayloadMapKeysShouldReturnEmptySliceWhenGivenEmptyMap(t *testing.T) {
	payload := make(map[string]interface{})
	values := getPayloadMapKeys(payload)
	require.Equal(t, []string{}, values)
}

func TestGetPayloadMapKeysShouldReturnEmptySliceWhenGivenNil(t *testing.T) {
	values := getPayloadMapKeys(nil)
	require.Equal(t, []string{}, values)
}

func TestGetPayloadMapKeysShouldReturnSliceWithOnlyRootKeys(t *testing.T) {
	payload := make(map[string]interface{})
	msg := `{
		"root_key_1": {
			"sub_key_1": "value1"
		}
	}`
	require.NoError(t, json.Unmarshal([]byte(msg), &payload))
	values := getPayloadMapKeys(payload)
	require.Equal(t, []string{"root_key_1"}, values)
}

func TestGetValueFromLogsShouldReturnEmptyStringWithAllNilArguments(t *testing.T) {
	require.Equal(t, "", utils.GetValueFromLogs(0, nil, "", ""))
}

func TestGetValueFromLogsShouldReturnEmptyStringIfValueNotFound(t *testing.T) {
	var logs types.ABCIMessageLogs
	logsJSON := `[
		{
			"msg_index": 0,
			"log": "",
			"events": [
				{
					"type": "test",
					"attributes": [
						{
							"key": "key1",
							"value": "val"
						}
					]
				}
			]
		}
	]`
	require.NoError(t, json.Unmarshal([]byte(logsJSON), &logs))
	require.Equal(t, "", utils.GetValueFromLogs(0, logs, "test", "key"))
}

func TestGetValueFromLogsShouldReturnCorrectValueWhenFound(t *testing.T) {
	var logs types.ABCIMessageLogs
	logsJSON := `[
		{
			"msg_index": 0,
			"log": "",
			"events": [
				{
					"type": "test",
					"attributes": [
						{
							"key": "key1",
							"value": "val"
						}
					]
				}
			]
		}
	]`
	require.NoError(t, json.Unmarshal([]byte(logsJSON), &logs))
	require.Equal(t, "val", utils.GetValueFromLogs(0, logs, "test", "key1"))
}
