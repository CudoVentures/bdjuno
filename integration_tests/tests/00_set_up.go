package test

import (
	"fmt"

	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
)

var (
	Name            = "name"
	Data            = "data"
	Minter          = "minter"
	Symbol          = "symbol"
	NotEditable     = "NotEditable"
	Schema          = "schema"
	Traits          = "traits"
	Description     = "description"
	Recipient       = "recipient"
	Metadata        = "metadata"
	URI             = "uri"
	MintRoyalties   = "mint-royalties"
	ResaleRoyalties = "resale-royalties"
	String100       = "100"
	String1         = "1"
	NoOwnerString   = "0x0"
	User1           = config.GetTestUserAddress(1)
	User2           = config.GetTestUserAddress(2)
	User3           = config.GetTestUserAddress(3)
	CudosAdmin      = config.GetTestAdminAddress()
	Royalties       = fmt.Sprintf("%s:%s", User2, String100)
)
