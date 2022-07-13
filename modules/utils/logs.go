package utils

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
