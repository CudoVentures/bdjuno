package gravity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isAttestationConsensusReached(t *testing.T) {
	type input struct {
		votes              int
		orchestratorsCount int
		result             bool
	}
	inputs := []input{
		{
			votes:              1,
			orchestratorsCount: 10,
			result:             false,
		},
		{
			votes:              1,
			orchestratorsCount: 1,
			result:             true,
		},
		{
			votes:              0,
			orchestratorsCount: 1,
			result:             false,
		},
		{
			votes:              0,
			orchestratorsCount: 0,
			result:             false,
		},
		{
			votes:              6,
			orchestratorsCount: 10,
			result:             true,
		},
		{
			votes:              5,
			orchestratorsCount: 10,
			result:             false,
		},
	}

	for i := 0; i < len(inputs); i++ {
		require.Equal(t, inputs[i].result, isConsensusReached(inputs[i].votes, inputs[i].orchestratorsCount))
	}
}
