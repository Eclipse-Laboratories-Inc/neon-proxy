package source

type Block struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       int    `json:"blockHeight"`
		BlockTime         int    `json:"blockTime"`
		Blockhash         string `json:"blockhash"`
		ParentSlot        int    `json:"parentSlot"`
		PreviousBlockhash string `json:"previousBlockhash"`
		Transactions      []Transaction
	} `json:"result"`
	ID    int `json:"id"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type TransactionSignatures struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		BlockTime          int         `json:"blockTime"`
		ConfirmationStatus string      `json:"confirmationStatus"`
		Err                interface{} `json:"err"`
		Memo               interface{} `json:"memo"`
		Signature          string      `json:"signature"`
		Slot               int         `json:"slot"`
	} `json:"result"`
	ID    int `json:"id"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Transaction struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BlockTime int `json:"blockTime"`
		Meta      struct {
			ComputeUnitsConsumed int         `json:"computeUnitsConsumed"`
			Err                  interface{} `json:"err"`
			Fee                  int         `json:"fee"`
			InnerInstructions    []struct {
				Index        int `json:"index"`
				Instructions []struct {
					Accounts       []int  `json:"accounts"`
					Data           string `json:"data"`
					ProgramIDIndex int    `json:"programIdIndex"`
				} `json:"instructions"`
			} `json:"innerInstructions"`
			LoadedAddresses struct {
				Readonly []interface{} `json:"readonly"`
				Writable []interface{} `json:"writable"`
			} `json:"loadedAddresses"`
			LogMessages       []string      `json:"logMessages"`
			PostBalances      []interface{} `json:"postBalances"`
			PostTokenBalances []struct {
				AccountIndex  int    `json:"accountIndex"`
				Mint          string `json:"mint"`
				Owner         string `json:"owner"`
				ProgramID     string `json:"programId"`
				UITokenAmount struct {
					Amount         string  `json:"amount"`
					Decimals       int     `json:"decimals"`
					UIAmount       float64 `json:"uiAmount"`
					UIAmountString string  `json:"uiAmountString"`
				} `json:"uiTokenAmount"`
			} `json:"postTokenBalances"`
			PreBalances      []interface{} `json:"preBalances"`
			PreTokenBalances []struct {
				AccountIndex  int    `json:"accountIndex"`
				Mint          string `json:"mint"`
				Owner         string `json:"owner"`
				ProgramID     string `json:"programId"`
				UITokenAmount struct {
					Amount         string  `json:"amount"`
					Decimals       int     `json:"decimals"`
					UIAmount       float64 `json:"uiAmount"`
					UIAmountString string  `json:"uiAmountString"`
				} `json:"uiTokenAmount"`
			} `json:"preTokenBalances"`
			Rewards []interface{} `json:"rewards"`
			Status  struct {
				Ok interface{} `json:"Ok"`
			} `json:"status"`
		} `json:"meta"`
		Slot        int `json:"slot"`
		Transaction struct {
			Message struct {
				AccountKeys []string `json:"accountKeys"`
				Header      struct {
					NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
					NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
					NumRequiredSignatures       int `json:"numRequiredSignatures"`
				} `json:"header"`
				Instructions []struct {
					Accounts       []interface{} `json:"accounts"`
					Data           string        `json:"data"`
					ProgramIDIndex int           `json:"programIdIndex"`
				} `json:"instructions"`
				RecentBlockhash string `json:"recentBlockhash"`
			} `json:"message"`
			Signatures []string `json:"signatures"`
		} `json:"transaction"`
	} `json:"result"`
	ID    int `json:"id"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// eth event log data parsed from solaan logs
type NeonLogTxEvent struct {
	eventType       int
	usedGas         int
	transactionHash []byte
	address         []byte
	topicList       []string
	data            []byte
}

// current log and transaction index
type Indexes struct {
	processed        map[string]int
	transactionIndex int
	logIndex         int
}

// eth event structure
type Event struct {
	TransactionHash string   `json:"transactionHash"`
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	LogIndex        string   `json:"logIndex"`
}

// eth tx block related params
type BlockParams struct {
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
	BlockHash        string `json:"blockHash"`
}

// defines eth log structure for each transaction
type EthLog struct {
	BlockParams
	Event
	Removed bool `json:"removed"`
}

type LogData struct {
	usedGas int
	logs    []string
}

type IterationCache struct {
	totalUsedGas   int
	targetTotalGas int
	logIterations  []LogData
}

// gas usage info parsed from logs
type NeonLogTxIx struct {
	gasUsed      int
	totalGasUsed int
}

// log return status
type NeonLogTxReturn struct {
	gasUsed int
	status  int
}
