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
	ResultCodeID          string
}

func NewMsgStoreCodeData(txHash, sender string, index int, success bool, instantiatePermission, resultCodeID string) *MsgStoreCodeData {
	return &MsgStoreCodeData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		InstantiatePermission: instantiatePermission,
		ResultCodeID:          resultCodeID,
	}
}

type MsgInstantiateContractData struct {
	CosmwasmMessageData

	Admin                 string
	Funds                 string
	Label                 string
	CodeID                string
	ResultContractAddress string
}

func NewMsgInstantiateContractData(txHash, sender string, index int, success bool, admin, funds, label, codeID, resultContractAddress string) *MsgInstantiateContractData {
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
		CodeID:                codeID,
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
	CodeID    string
	Arguments string
}

func NewMsgMigrateContractData(txHash, sender string, index int, success bool, contract, codeID, arguments string) *MsgMigrateContractData {
	return &MsgMigrateContractData{
		CosmwasmMessageData: CosmwasmMessageData{
			TxHash:  txHash,
			Index:   index,
			Sender:  sender,
			Success: success,
		},
		Contract:  contract,
		CodeID:    codeID,
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
