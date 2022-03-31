package types

type CosmwasmMessageData struct {
	TxHash  string
	Index   int
	Sender  string
	Success bool
}

type MsgStoreCodeData struct {
	CosmwasmMessageData

	InstantiatePermission string
	ResultCodeId          string
}

func NewMsgStoreCodeData(txHash, sender string, index int, success bool, instantiatePermission, resultCodeId string) *MsgStoreCodeData {
	return &MsgStoreCodeData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		InstantiatePermission: instantiatePermission,
		ResultCodeId:          resultCodeId,
	}
}

type MsgInstantiateContractData struct {
	CosmwasmMessageData

	Admin                 string
	Funds                 string
	Label                 string
	CodeId                string
	ResultContractAddress string
}

func NewMsgInstantiateContractData(txHash, sender string, index int, success bool, admin, funds, label, codeId, resultContractAddress string) *MsgInstantiateContractData {
	return &MsgInstantiateContractData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Admin:                 admin,
		Funds:                 funds,
		Label:                 label,
		CodeId:                codeId,
		ResultContractAddress: resultContractAddress,
	}
}

type MsgExecuteContractData struct {
	CosmwasmMessageData

	Method    string
	Arguments string
	Funds     string
	Contract  string
}

func NewMsgExecuteContractData(txHash, sender string, index int, success bool, method, arguments, funds, contract string) *MsgExecuteContractData {
	return &MsgExecuteContractData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Method:    method,
		Arguments: arguments,
		Funds:     funds,
		Contract:  contract,
	}
}

type MsgMigrateContractData struct {
	CosmwasmMessageData

	Contract  string
	CodeId    string
	Arguments string
}

func NewMsgMigrateContractData(txHash, sender string, index int, success bool, contract, codeId, arguments string) *MsgMigrateContractData {
	return &MsgMigrateContractData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Contract:  contract,
		CodeId:    codeId,
		Arguments: arguments,
	}
}

type MsgUpdateAdminData struct {
	CosmwasmMessageData

	Contract string
	NewAdmin string
}

func NewMsgUpdateAdminData(txHash, sender string, index int, success bool, contract, newAdmin string) *MsgUpdateAdminData {
	return &MsgUpdateAdminData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Contract: contract,
		NewAdmin: newAdmin,
	}
}

type MsgClearAdminData struct {
	CosmwasmMessageData

	Contract string
}

func NewClearAdminData(txHash, sender string, index int, success bool, contract string) *MsgClearAdminData {
	return &MsgClearAdminData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Contract: contract,
	}
}
