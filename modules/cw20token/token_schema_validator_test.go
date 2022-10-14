package cw20token

import (
	"fmt"
	"testing"

	"github.com/forbole/bdjuno/v2/types"
	"github.com/stretchr/testify/require"
)

func TestCw20Token_ValidateTokenSchema(t *testing.T) {
	for _, tc := range []struct {
		name    string
		msg     *types.MsgVerifiedContract
		wantErr error
	}{
		{
			name: "happy path",
			msg:  newPubMsg(1, validExecuteSchema, validQuerySchema),
		},
		{
			name:    "invalid query schema",
			msg:     newPubMsg(1, validExecuteSchema, invalidQuerySchema),
			wantErr: fmt.Errorf("(root): Must validate one and only one schema (oneOf)\n(root): balance is required\n(root): Additional property token_info is not allowed\n"),
		},
		{
			name:    "invalid execute schema",
			msg:     newPubMsg(1, invalidExecuteSchema, validQuerySchema),
			wantErr: fmt.Errorf("(root): Must validate one and only one schema (oneOf)\n(root): transfer is required\n(root): Additional property send is not allowed\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTokenSchema(tc.msg)
			require.Equal(t, tc.wantErr, err)
		})
	}
}
