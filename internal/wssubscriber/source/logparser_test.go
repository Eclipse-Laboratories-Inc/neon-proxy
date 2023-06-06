package source

import (
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/config"
	"github.com/test-go/testify/assert"
	"os"
	"testing"
)

func TestSimpleLogs(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [50]string{
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke [1]",
		"Program log: Instruction: Execute Transaction from Instruction",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program data: RU5URVI= Q0FMTA== SR/8buQv77Ttq5un1fPmOZWeCBs=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: TE9HMw== qiSlpeJz76pkqWCyjebie4f/3/w= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsI=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMg== 8QQVltoEmcNDjjset7lTVMau0fU= Ag== 4f/8xJI9BLVZ9NKai/xs2gTrWw08RgdRwkAsXFzJEJw= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMw== 8QQVltoEmcNDjjset7lTVMau0fU= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAVO6SN3pjxTojcDdA3n7qCd/W7hE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFruyM=",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN3Zb/1Y=",
		"Program data: TE9HMQ== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= AQ== HEEempbgcSQcLyH3cmsXronjyrTHi+UOBisDqf/7utE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAed6m/4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAHO0Xvd0KAMWLGA==",
		"Program data: TE9HMg== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Ag== TCCbX8itUHWPE+LhCIulalYN/2kKHG/vJjlPTAOCHE8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKxxxpO8I5BQ==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== e05WMSXjzzae6aRWU+vIVS0/x9s=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
		"Program log: Instruction: Transfer",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 182186 compute units",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program log: exit_status=0x12",
		"Program data: UkVUVVJO Eg==",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU consumed 1241647 of 1399944 compute units",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"}

	// get events from logs
	events, err := GetEvents(logMessages[:])
	if err != nil {
		t.Errorf("processing logs got error %s", err.Error())
	}

	// check num of events
	if len(events) != 7 {
		t.Errorf("Incorrect number of events, must be 7")
	}

	if events[0].Address != "0xaa24a5a5e273efaa64a960b28de6e27b87ffdffc" {
		t.Errorf("Event address not correct")
	}

	if events[0].Data != "0x00000000000000000000000000000000000000000000000000000000000922c2" {
		t.Errorf("Event data not correct")
	}

	if events[0].Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" || events[0].Topics[1] != "0x0000000000000000000000007b4e563125e3cf369ee9a45653ebc8552d3fc7db" || events[0].Topics[2] != "0x000000000000000000000000129399072d8b834b6f4c541f907e0ca20ad09944" {
		t.Errorf("event topics not correct")
	}

	if events[0].TransactionHash != "0xae9bc48d9c011495caa38651c8a2787403a823084f608a794af9eb4eebbabf5c" {
		t.Errorf("event tx hash not correct")
	}

	if events[6].Address != "0x129399072d8b834b6f4c541f907e0ca20ad09944" {
		t.Errorf("Event address not correct")
	}

	if events[6].Data != "0x00000000000000000000000000000000000000000000000000000000000922c20000000000000000000000000000000000000000000000008ac71c693bc23905" {
		t.Errorf("Event data not correct")
	}

	if events[6].Topics[0] != "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f" || events[6].Topics[1] != "0x000000000000000000000000491ffc6ee42fefb4edab9ba7d5f3e639959e081b" {
		t.Errorf("event topics not correct %s %s", events[6].Topics[0], events[6].Topics[1])
	}

	if events[6].TransactionHash != "0xae9bc48d9c011495caa38651c8a2787403a823084f608a794af9eb4eebbabf5c" {
		t.Errorf("event tx hash not correct")
	}
}

func TestDoubleTxHash(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [50]string{
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke [1]",
		"Program log: Instruction: Execute Transaction from Instruction",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program data: RU5URVI= Q0FMTA== SR/8buQv77Ttq5un1fPmOZWeCBs=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: TE9HMw== qiSlpeJz76pkqWCyjebie4f/3/w= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsI=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMg== 8QQVltoEmcNDjjset7lTVMau0fU= Ag== 4f/8xJI9BLVZ9NKai/xs2gTrWw08RgdRwkAsXFzJEJw= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMw== 8QQVltoEmcNDjjset7lTVMau0fU= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAVO6SN3pjxTojcDdA3n7qCd/W7hE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFruyM=",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN3Zb/1Y=",
		"Program data: TE9HMQ== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= AQ== HEEempbgcSQcLyH3cmsXronjyrTHi+UOBisDqf/7utE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAed6m/4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAHO0Xvd0KAMWLGA==",
		"Program data: TE9HMg== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Ag== TCCbX8itUHWPE+LhCIulalYN/2kKHG/vJjlPTAOCHE8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKxxxpO8I5BQ==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== e05WMSXjzzae6aRWU+vIVS0/x9s=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
		"Program log: Instruction: Transfer",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 182186 compute units",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program log: exit_status=0x12",
		"Program data: UkVUVVJO Eg==",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU consumed 1241647 of 1399944 compute units",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"}

	// get events from logs
	_, err = GetEvents(logMessages[:])
	if err != nil {
		t.Errorf("must not have error here for having tx hash twice in the logs")
	}
}

func TestUnfinishedEnterCall(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [60]string{
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke [1]",
		"Program log: Instruction: Execute Transaction from Instruction",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program data: RU5URVI= Q0FMTA== SR/8buQv77Ttq5un1fPmOZWeCBs=",
		"Program data: RU5URVI= Q0FMTA== SR/8buQv77Ttq5un1fPmOZWeCBs=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: TE9HMw== qiSlpeJz76pkqWCyjebie4f/3/w= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsI=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMg== 8QQVltoEmcNDjjset7lTVMau0fU= Ag== 4f/8xJI9BLVZ9NKai/xs2gTrWw08RgdRwkAsXFzJEJw= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMw== 8QQVltoEmcNDjjset7lTVMau0fU= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAVO6SN3pjxTojcDdA3n7qCd/W7hE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFruyM=",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN3Zb/1Y=",
		"Program data: TE9HMQ== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= AQ== HEEempbgcSQcLyH3cmsXronjyrTHi+UOBisDqf/7utE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAed6m/4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAHO0Xvd0KAMWLGA==",
		"Program data: TE9HMg== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Ag== TCCbX8itUHWPE+LhCIulalYN/2kKHG/vJjlPTAOCHE8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKxxxpO8I5BQ==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== e05WMSXjzzae6aRWU+vIVS0/x9s=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
		"Program log: Instruction: Transfer",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 182186 compute units",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program log: exit_status=0x12",
		"Program data: UkVUVVJO Eg==",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU consumed 1241647 of 1399944 compute units",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"}

	// get events from logs
	_, err = GetEvents(logMessages[:])
	if err == nil {
		t.Errorf("processing logs must have error for unfinished evm function calls")
	}
}

func TestSkippingUnfinishedTx(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [5]string{
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke [1]",
		"Program log: Instruction: Write To Holder",
		"Program data: SEFTSA== LjyfNnLCzXkN38xkBWIcJdsk37YE56M5y7g4tKF9mpo=",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU consumed 4008 of 200000 compute units",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"}

	// get events from logs
	events, err := GetEvents(logMessages[:])
	if err != nil {
		t.Errorf("processing logs got error %s", err.Error())
	}

	if len(events) != 0 {
		t.Errorf("event num should be 0 have %d", len(events))
	}
}

func TestManyGasOuts(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "ComputeBudget111111111111111111111111111111")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [60]string{"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke [1]",
		"Program log: Instruction: Execute Transaction from Instruction",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program data: SEFTSA== rpvEjZwBFJXKo4ZRyKJ4dAOoIwhPYIp5SvnrTuu6v1w=",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program data: RU5URVI= Q0FMTA== SR/8buQv77Ttq5un1fPmOZWeCBs=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: TE9HMw== qiSlpeJz76pkqWCyjebie4f/3/w= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsI=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMg== 8QQVltoEmcNDjjset7lTVMau0fU= Ag== 4f/8xJI9BLVZ9NKai/xs2gTrWw08RgdRwkAsXFzJEJw= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RU5URVI= Q0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: TE9HMw== 8QQVltoEmcNDjjset7lTVMau0fU= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAEpOZBy2Lg0tvTFQfkH4MogrQmUQ= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAisccaTvCOQU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== EpOZBy2Lg0tvTFQfkH4MogrQmUQ=",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== qiSlpeJz76pkqWCyjebie4f/3/w=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== 8QQVltoEmcNDjjset7lTVMau0fU=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= U1RBVElDQ0FMTA== aW1z1yYiI3JNYLLOnW4g/DHfxWs=",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAVO6SN3pjxTojcDdA3n7qCd/W7hE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFruyM=",
		"Program data: TE9HMw== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Aw== 3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs+8= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= AAAAAAAAAAAAAAAAe05WMSXjzzae6aRWU+vIVS0/x9s= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAN3Zb/1Y=",
		"Program data: TE9HMQ== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= AQ== HEEempbgcSQcLyH3cmsXronjyrTHi+UOBisDqf/7utE= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAed6m/4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAHO0Xvd0KAMWLGA==",
		"Program data: TE9HMg== EpOZBy2Lg0tvTFQfkH4MogrQmUQ= Ag== TCCbX8itUHWPE+LhCIulalYN/2kKHG/vJjlPTAOCHE8= AAAAAAAAAAAAAAAASR/8buQv77Ttq5un1fPmOZWeCBs= AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJIsIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKxxxpO8I5BQ==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program data: RU5URVI= Q0FMTA== e05WMSXjzzae6aRWU+vIVS0/x9s=",
		"Program data: RVhJVA== U1RPUA==",
		"Program data: RVhJVA== UkVUVVJO",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
		"Program log: Instruction: Transfer",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 182186 compute units",
		"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program log: exit_status=0x12",
		"Program data: UkVUVVJO Eg==",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU consumed 1241647 of 1399944 compute units",
		"Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"}

	// get events from logs
	_, err = GetEvents(logMessages[:])
	if err != nil {
		t.Errorf("must not have error here for having tx hash twice in the logs")
	}
}

// test checks bug, detected 6.06.2023 NDEV-1754
func TestInvokeWithoutSuccess(t *testing.T) {
	err := os.Setenv("EVM_ADDRESS", "ComputeBudget111111111111111111111111111111")
	assert.Empty(t, err)

	EvmInvocationLog = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd = "Program " + os.Getenv(config.EvmAddress) + " fail"

	logMessages := [7]string{
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"Program data: R0FT ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}

	// get events from logs
	_, err = GetEvents(logMessages[:])
	if err != nil {
		t.Errorf("must not have error here for having tx hash twice in the logs")
	}
}
