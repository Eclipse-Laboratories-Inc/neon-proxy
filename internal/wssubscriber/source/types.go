package source

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Header struct {
	NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
	NumRequiredSignatures       int `json:"numRequiredSignatures"`
}

type Instruction struct {
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	ProgramIdIndex int    `json:"programIdIndex"`
}

type Message struct {
	AccountKeys     []string      `json:"accountKeys"`
	Header          Header        `json:"header"`
	Instructions    []Instruction `json:"instructions"`
	RecentBlockhash string        `json:"recentBlockhash"`
}

type Transaction struct {
	Message    Message  `json:"message"`
	Signatures []string `json:"signatures"`
}

type TransactionFull struct {
	Meta        Meta
	Transaction Transaction
}

type Status struct {
	Ok interface{} `json:"Ok"`
}

type Meta struct {
	Err               interface{}   `json:"err"`
	Fee               int           `json:"fee"`
	InnerInstructions []interface{} `json:"innerInstructions"`
	LogMessages       []string      `json:"logMessages"`
	PostBalances      []int64       `json:"postBalances"`
	PostTokenBalances []interface{} `json:"postTokenBalances"`
	PreBalances       []int64       `json:"preBalances"`
	PreTokenBalances  []interface{} `json:"preTokenBalances"`
	Rewards           interface{}   `json:"rewards"`
	Status            Status        `json:"status"`
}

type Block struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       int               `json:"blockHeight"`
		BlockTime         interface{}       `json:"blockTime"`
		Blockhash         string            `json:"blockhash"`
		ParentSlot        int               `json:"parentSlot"`
		PreviousBlockhash string            `json:"previousBlockhash"`
		Transactions      []TransactionFull `json:"transactions"`
	} `json:"result"`
	Error *Error `json:"error,omitempty"`
	Id    int    `json:"id"`
}

type BlockSlot struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  uint64 `json:"result"`
	ID      int    `json:"id"`
	Error   Error  `json:"error,omitempty"`
}

type GetTransactionSignatureByAccountKeyResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   *Error `json:"error,omitempty"`
	Result  []struct {
		Memo      interface{} `json:"memo"`
		Signature string      `json:"signature"`
		Slot      int64       `json:"slot"`
		BlockTime int64       `json:"blockTime"`
	} `json:"result"`
	Id int `json:"id"`
}

type Log struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}
