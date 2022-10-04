package cw20token

import (
	"fmt"

	"github.com/forbole/bdjuno/v2/types"
	gjv "github.com/xeipuuv/gojsonschema"
)

func isToken(contract *types.VerifiedContractPublishMessage) (bool, error) {
	executeMsgs := []string{
		`{"transfer":{"recipient":"test","amount":"1"}}`,
		`{"send":{"contract":"test","amount":"1","msg":"test"}}`,
	}
	if err := validateSchema(contract.ExecuteSchema, executeMsgs); err != nil {
		return false, err
	}

	queryMsgs := []string{
		`{"balance":{"address":"test"}}`,
		`{"token_info":{}}`,
		`{"all_accounts":{}}`,
	}
	if err := validateSchema(contract.QuerySchema, queryMsgs); err != nil {
		return false, err
	}

	return true, nil
}

func validateSchema(schema string, msgs []string) error {
	for _, msg := range msgs {
		if result, err := gjv.Validate(gjv.NewStringLoader(schema), gjv.NewStringLoader(msg)); err != nil {
			return err
		} else if !result.Valid() {
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
