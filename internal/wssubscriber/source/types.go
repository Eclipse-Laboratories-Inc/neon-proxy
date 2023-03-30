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

type Transaction struct {
	Meta struct {
		ComputeUnitsConsumed int           `json:"computeUnitsConsumed"`
		Err                  interface{}   `json:"err"`
		Fee                  int           `json:"fee"`
		InnerInstructions    []interface{} `json:"innerInstructions"`
		LoadedAddresses      struct {
			Readonly []interface{} `json:"readonly"`
			Writable []interface{} `json:"writable"`
		} `json:"loadedAddresses"`
		LogMessages       []string      `json:"logMessages"`
		PostBalances      []interface{} `json:"postBalances"`
		PostTokenBalances []interface{} `json:"postTokenBalances"`
		PreBalances       []interface{} `json:"preBalances"`
		PreTokenBalances  []interface{} `json:"preTokenBalances"`
		Rewards           interface{}   `json:"rewards"`
		Status            struct {
			Ok interface{} `json:"Ok"`
		} `json:"status"`
	} `json:"meta"`
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
}
