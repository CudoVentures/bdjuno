package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v3/node"

	"github.com/forbole/bdjuno/v2/modules"
)

// Context contains the data about a Hasura actions worker execution
type Context struct {
	Node    node.Node
	Sources *modules.Sources
	Cdc     codec.Codec
}

// NewContext returns a new Context instance
func NewContext(node node.Node, sources *modules.Sources, cdc codec.Codec) *Context {
	return &Context{
		Node:    node,
		Sources: sources,
		Cdc:     cdc,
	}
}

// GetHeight uses the lastest height when the input height is empty from graphql request
func (c *Context) GetHeight(payload *Payload) (int64, error) {
	if payload == nil || payload.Input.Height == 0 {
		latestHeight, err := c.Node.LatestHeight()
		if err != nil {
			return 0, fmt.Errorf("error while getting chain latest block height: %s", err)
		}
		return latestHeight, nil
	}

	return payload.Input.Height, nil
}

// ActionHandler represents a Hasura action request handler.
// It returns an interface to be returned to the called, or an error if something is wrong
type ActionHandler = func(context *Context, payload *Payload) (interface{}, error)
type NftTransferEventsActionHandler = func(context *Context, payload *NftTransferEventsPayload) (interface{}, error)
