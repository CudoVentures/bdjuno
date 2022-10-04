package cw20token

import (
	"fmt"
	"testing"

	"github.com/forbole/bdjuno/v2/types"
	"github.com/stretchr/testify/require"
)

func TestCw20Token_IsToken(t *testing.T) {
	for _, tc := range []struct {
		name    string
		req     *types.VerifiedContractPublishMessage
		want    bool
		wantErr error
	}{
		{
			name:    "valid schemas",
			req:     newPubMsg(1, validExecuteSchema, validQuerySchema),
			want:    true,
			wantErr: nil,
		},
		{
			name:    "invalid execute schema",
			req:     newPubMsg(1, invalidExecuteSchema, validQuerySchema),
			want:    false,
			wantErr: fmt.Errorf("(root): Must validate one and only one schema (oneOf)\n(root): transfer is required\n(root): Additional property send is not allowed\n"),
		},
		{
			name:    "invalid query schema",
			req:     newPubMsg(1, validExecuteSchema, invalidQuerySchema),
			want:    false,
			wantErr: fmt.Errorf("(root): Must validate one and only one schema (oneOf)\n(root): balance is required\n(root): Additional property token_info is not allowed\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have, haveErr := isToken(tc.req)
			require.Equal(t, tc.want, have)
			require.Equal(t, tc.wantErr, haveErr)
		})
	}
}

const (
	validExecuteSchema   = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"Cw20ExecuteMsg","oneOf":[{"type":"object","required":["transfer"],"properties":{"transfer":{"type":"object","required":["amount","recipient"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"recipient":{"type":"string"}}}},"additionalProperties":false},{"type":"object","required":["send"],"properties":{"send":{"type":"object","required":["amount","contract","msg"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"contract":{"type":"string"},"msg":{"$ref":"#/definitions/Binary"}}}},"additionalProperties":false}],"definitions":{"Uint128":{"type":"string"},"Uint64":{"type":"string"},"Binary":{"type":"string"}}}`
	validQuerySchema     = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"QueryMsg","oneOf":[{"type":"object","required":["balance"],"properties":{"balance":{"type":"object","required":["address"],"properties":{"address":{"type":"string"}}}},"additionalProperties":false},{"type":"object","required":["token_info"],"properties":{"token_info":{"type":"object"}},"additionalProperties":false},{"type":"object","required":["all_accounts"],"properties":{"all_accounts":{"type":"object","properties":{"limit":{"type":["integer","null"],"format":"uint32","minimum":0},"start_after":{"type":["string","null"]}}}},"additionalProperties":false}]}`
	invalidExecuteSchema = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"Cw20ExecuteMsg","oneOf":[{"type":"object","required":["transfer"],"properties":{"transfer":{"type":"object","required":["amount","recipient"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"recipient":{"type":"string"}}}},"additionalProperties":false}],"definitions":{"Uint128":{"type":"string"}}}`
	invalidQuerySchema   = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"QueryMsg","oneOf":[{"type":"object","required":["balance"],"properties":{"balance":{"type":"object","required":["address"],"properties":{"address":{"type":"string"}}}},"additionalProperties":false}]}`
)
