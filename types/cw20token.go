package types

import "encoding/json"

type Mint struct {
	Minter    string `json:"minter"`
	MaxSupply uint64 `json:"cap,string"`
}

type Marketing struct {
	Project     string          `json:"project"`
	Description string          `json:"description"`
	Admin       string          `json:"marketing"`
	Logo        json.RawMessage `json:"logo"`
}

type TokenBalance struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount,string"`
}

type TokenInfo struct {
	Address     string         `json:"address,omitempty"`
	Name        string         `json:"name"`
	Symbol      string         `json:"symbol"`
	Decimals    int8           `json:"decimals"`
	TotalSupply uint64         `json:"total_supply,string"`
	Mint        Mint           `json:"mint,omitempty"`
	Marketing   Marketing      `json:"marketing_info,omitempty"`
	CodeID      uint64         `json:"code_id"`
	Balances    []TokenBalance `json:"initial_balances,omitempty"`
}

type MsgExecuteToken struct {
	Type              string
	MsgRaw            []byte
	Contract          string
	Sender            string
	Recipient         string
	Amount            uint64 `json:"amount,string"`
	Owner             string
	RecipientContract string `json:"contract"`
	NewMinter         string `json:"new_minter"`
	Project           string
	Description       string
	MarketingAdmin    string `json:"marketing"`
}

type TypeMsgExecute string

const (
	TypeTransfer        TypeMsgExecute = "transfer"
	TypeTransferFrom    TypeMsgExecute = "transfer_from"
	TypeSend            TypeMsgExecute = "send"
	TypeSendFrom        TypeMsgExecute = "send_from"
	TypeBurn            TypeMsgExecute = "burn"
	TypeBurnFrom        TypeMsgExecute = "burn_from"
	TypeMint            TypeMsgExecute = "mint"
	TypeUpdateMinter    TypeMsgExecute = "update_minter"
	TypeUpdateMarketing TypeMsgExecute = "update_marketing"
	TypeUploadLogo      TypeMsgExecute = "upload_logo"
)

type MsgTransfer struct {
	Recipient string
	Amount    uint64 `json:"amount,string"`
}

type MsgTransferFrom struct {
	Owner     string
	Recipient string
	Amount    uint64 `json:"amount,string"`
}

type MsgBurn struct {
	Amount uint64 `json:"amount,string"`
}

type MsgBurnFrom struct {
	Owner  string
	Amount uint64 `json:"amount,string"`
}

type MsgSend struct {
	Contract string
	Amount   uint64 `json:"amount,string"`
	Msg      json.RawMessage
}

type MsgSendFrom struct {
	Owner    string
	Contract string
	Amount   uint64 `json:"amount,string"`
	Msg      json.RawMessage
}

type MsgMint struct {
	Recipient string
	Amount    uint64 `json:"amount,string"`
}

type MsgUpdateMinter struct {
	NewMinter string `json:"new_minter"`
}
type MsgUpdateMarketing struct {
	Project     string `json:"project"`
	Description string `json:"description"`
	Admin       string `json:"marketing"`
}

type MsgUploadLogo json.RawMessage

type MsgExecute struct {
	Transfer        MsgTransfer        `json:"transfer"`
	TransferFrom    MsgTransferFrom    `json:"transfer_from"`
	Send            MsgSend            `json:"send"`
	SendFrom        MsgSendFrom        `json:"send_from"`
	Burn            MsgBurn            `json:"burn"`
	BurnFrom        MsgBurnFrom        `json:"burn_from"`
	Mint            MsgMint            `json:"mint"`
	UpdateMinter    MsgUpdateMinter    `json:"update_minter"`
	UpdateMarketing MsgUpdateMarketing `json:"update_marketing"`
	UploadLogo      json.RawMessage    `json:"upload_logo"`
}
