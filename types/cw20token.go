package types

import "encoding/json"

type Mint struct {
	Minter    string
	MaxSupply string `json:"cap"`
}

type Marketing struct {
	Project     string
	Description string
	Admin       string `json:"marketing"`
	Logo        *json.RawMessage
}

type TokenBalance struct {
	Address string
	Amount  string
}

type TokenInfo struct {
	Name          string
	Symbol        string
	Decimals      int8
	Address       string         `json:"address,omitempty"`
	TotalSupply   string         `json:"total_supply"`
	Mint          Mint           `json:"mint,omitempty"`
	Marketing     Marketing      `json:"marketing_info,omitempty"`
	CodeID        uint64         `json:"code_id"`
	Balances      []TokenBalance `json:"initial_balances,omitempty"`
	InitialSupply string
	Type          string
	Creator       string
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
	Amount    string
}

type MsgTransferFrom struct {
	Owner     string
	Recipient string
	Amount    string
}

type MsgBurn struct {
	Amount string
}

type MsgBurnFrom struct {
	Owner  string
	Amount string
}

type MsgSend struct {
	Contract string
	Amount   string
	Msg      json.RawMessage
}

type MsgSendFrom struct {
	Owner    string
	Contract string
	Amount   string
	Msg      json.RawMessage
}

type MsgMint struct {
	Recipient string
	Amount    string
}

type MsgUpdateMinter struct {
	NewMinter string `json:"new_minter"`
}
type MsgUpdateMarketing struct {
	Project     string
	Description string
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
