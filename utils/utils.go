package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/metadata"
)

// RemoveDuplicateValues removes the duplicated values from the given slice
func RemoveDuplicateValues(slice []string) []string {
	keys := make(map[string]bool)
	var list []string

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// GetHeightRequestContext adds the height to the context for queries
func GetHeightRequestContext(context context.Context, height int64) context.Context {
	return metadata.AppendToOutgoingContext(
		context,
		grpctypes.GRPCBlockHeightHeader,
		strconv.FormatInt(height, 10),
	)
}

func SanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func GetData(data string) (string, string) {
	dataText := data
	dataJSON := "{}"

	if data != "" && IsJSON(data) {
		dataJSON = data
		dataText = ""
	}

	return dataJSON, dataText
}

func GetValueFromLogs(index uint32, logs sdk.ABCIMessageLogs, eventType, attributeKey string) string {
	for _, log := range logs {
		if log.MsgIndex != index {
			continue
		}

		for _, event := range log.Events {
			if event.Type != eventType {
				continue
			}

			for _, attr := range event.Attributes {
				if attr.Key == attributeKey {
					return strings.ReplaceAll(attr.Value, "\"", "")
				}
			}
		}
	}

	return ""
}

func GetUint64FromLogs(index int, logs sdk.ABCIMessageLogs, txHash, eventType, attributeKey string) (uint64, error) {
	valueStr := GetValueFromLogs(uint32(index), logs, eventType, attributeKey)
	if valueStr == "" {
		return 0, fmt.Errorf("attribute %s for event %s not found in tx %s", attributeKey, eventType, txHash)
	}

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s from tx %s to uint64", valueStr, txHash)
	}

	return value, nil
}
