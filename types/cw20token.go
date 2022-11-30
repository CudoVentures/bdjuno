package types

import "encoding/json"

type Mint struct {
	Minter    string `json:"minter"`
	MaxSupply string `json:"cap"`
}

type Marketing struct {
	Project     string          `json:"project"`
	Description string          `json:"description"`
	Admin       string          `json:"marketing"`
	Logo        json.RawMessage `json:"logo"`
}

type TokenBalance struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type TokenInfo struct {
	Address       string         `json:"address,omitempty"`
	Name          string         `json:"name"`
	Symbol        string         `json:"symbol"`
	Decimals      int8           `json:"decimals"`
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
	Amount    string `json:"amount"`
}

type MsgTransferFrom struct {
	Owner     string
	Recipient string
	Amount    string `json:"amount"`
}

type MsgBurn struct {
	Amount string `json:"amount"`
}

type MsgBurnFrom struct {
	Owner  string
	Amount string `json:"amount"`
}

type MsgSend struct {
	Contract string
	Amount   string `json:"amount"`
	Msg      json.RawMessage
}

type MsgSendFrom struct {
	Owner    string
	Contract string
	Amount   string `json:"amount"`
	Msg      json.RawMessage
}

type MsgMint struct {
	Recipient string
	Amount    string `json:"amount"`
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
