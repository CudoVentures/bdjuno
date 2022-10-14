package cw20token

import (
	"fmt"

	"github.com/forbole/bdjuno/v2/types"
	gjv "github.com/xeipuuv/gojsonschema"
)

func validateTokenSchema(contract *types.MsgVerifiedContract) error {
	executeMsgs := []string{
		`{"transfer":{"recipient":"test","amount":"1"}}`,
		`{"send":{"contract":"test","amount":"1","msg":"test"}}`,
	}
	if err := validateSchema(contract.ExecuteSchema, executeMsgs); err != nil {
		return err
	}

	queryMsgs := []string{
		`{"balance":{"address":"test"}}`,
		`{"token_info":{}}`,
		`{"all_accounts":{}}`,
	}

	return validateSchema(contract.QuerySchema, queryMsgs)
}

func validateSchema(schema string, msgs []string) error {
	for _, msg := range msgs {
		result, err := gjv.Validate(gjv.NewStringLoader(schema), gjv.NewStringLoader(msg))
		if err != nil {
			return err
		}

		if !result.Valid() {
			err := ""
			for _, e := range result.Errors() {
				err += e.String()
				err += "\n"
			}
			return fmt.Errorf("%s", err)
		}
	}

	return nil
}
