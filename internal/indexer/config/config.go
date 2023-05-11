package config

import (
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/config"
	"github.com/shopspring/decimal"
	"math"
	"time"
)

const (
	mempoolCapacityMax             = 4096
	mempoolExecutorLimitCnt        = 1024
	mempoolCacheLifeSecMax         = 900 * time.Second
	holderSizeMax                  = 131072
	minOpBalanceToWarnMax          = 9000000000
	minOpBalanceToErrMax           = 1000000000
	permAccountIdMax               = 1
	permAccountLimitMax            = 2
	recheckUsedResourceSecMax      = 60 * time.Second
	recheckResourceAfterUsesCntMax = 60
	accountPermissionUpdateIntMax  = 300
	slotProcessingDelayMax         = 0 * time.Second
	minGasPriceMax                 = 1
	minWoChainidGasPriceMax        = 10
	neonDecimalsMax                = 9
	indexerParallelRequestCntMax   = 10
	indexerPollCntMax              = 1000
	indexerLogSkipCntMax           = 1000
	indexerCheckMsecMax            = 200 * time.Second
	maxTxAccountCntMax             = 62
	fuzzFailPctMax                 = 0
	confirmCheckMsecMax            = 100 * time.Millisecond
	maxEvmStepCntEmulateMax        = 500000
	genesisTimestampMax            = 0
	updatePythMappingPeriodSecMax  = 360 * time.Second
)

type IndexerConfig struct {
	OneBlockSec                 time.Duration      `json:"ONE_BLOCK_SEC"`
	MinFinalizeSec              time.Duration      `json:"MIN_FINALIZE_SEC"`
	SolanaEndpoint              string             `json:"SOLANA_URL" env:"SOLANA_URL"`
	PpSolanaEndpoint            string             `json:"PP_SOLANA_URL" env:"PP_SOLANA_URL"`
	EvmLoaderID                 solana.PublicKey   `json:"EVM_LOADER_ID"`
	MempoolCapacity             int                `json:"MEMPOOL_CAPACITY" default:"4"`
	MempoolExecutorLimitCnt     int                `json:"MEMPOOL_EXECUTOR_LIMIT_CNtime.DurationT" default:"4"`
	MempoolCacheLifeSec         time.Duration      `json:"MEMPOOL_CACHE_LIFE_SEC" default:"15s"`
	HolderSize                  int                `json:"HOLDER_SIZE" default:"1024"`
	MinOpBalanceToWarn          int                `json:"MIN_OP_BALANCE_TO_WARN" default:"9000000000"`
	MinOpBalanceToErr           int                `json:"MIN_OP_BALANCE_TO_ERR" default:"1000000000"`
	PermAccountId               int                `json:"PERM_ACCOUNT_ID" default:"1"`
	PermAccountLimit            int                `json:"PERM_ACCOUNT_LIMIT" default:"1"`
	RecheckUsedResourceSec      time.Duration      `json:"RECHECK_USED_RESOURCE_SEC" default:"10s"`
	RecheckResourceAfterUsesCnt int                `json:"RECHECK_RESOURCE_AFTER_USES_CNT" default:"10"`
	RetryOnFail                 int                `json:"RETRY_ON_FAIL" default:"1"`
	EnablePrivateApi            bool               `json:"ENABLE_PRIVATE_API" default:"false"`
	EnableSendTxApi             bool               `json:"ENABLE_SEND_TX_API" default:"true"`
	UseEarliestBlockIf0Passed   bool               `json:"USE_EARLIEST_BLOCK_IF_0_PASSED" default:"false"`
	AccountPermissionUpdateInt  int                `json:"ACCOUNT_PERMISSION_UPDATE_INT" default:"10" max:"300"`
	AllowUnderpricedTxWoChainid bool               `json:"ALLOW_UNDERPRICED_TX_WO_CHAINID" default:"false"`
	SlotProcessingDelay         time.Duration      `json:"SLOT_PROCESSING_DELAY" default:"0s" max:"0s"`
	ExtraGasPct                 decimal.Decimal    `json:"EXTRA_GAS_PCT" default:"0.0"`
	OperatorFee                 decimal.Decimal    `json:"OPERATOR_FEE" default:"0.1"`
	GasPriceSuggestedPct        decimal.Decimal    `json:"GAS_PRICE_SUGGESTED_PCT" default:"0.01"`
	MinGasPrice                 decimal.Decimal    `json:"MIN_GAS_PRICE" default:"0" max:"1"`
	MinWoChainidGasPrice        decimal.Decimal    `json:"MIN_WO_CHAINID_GAS_PRICE" default:"0" max:"10"`
	NeonPriceUsd                decimal.Decimal    `json:"NEON_PRICE_USD" default:"0.25"`
	NeonDecimals                decimal.Decimal    `json:"NEON_DECIMALS" default:"1" max:"9"`
	StartSlot                   string             `json:"START_SLOT" default:"'0'"`
	IndexerParallelRequestCnt   int                `json:"INDEXER_PARALLEL_REQUEST_CNT" default:"1" max:"10"`
	IndexerPollCnt              int                `json:"INDEXER_POLL_CNT" default:"1" max:"1000"`
	IndexerLogSkipCnt           int                `json:"INDEXER_LOG_SKIP_CNT" default:"1" max:"1000"`
	IndexerCheckMsec            time.Duration      `json:"INDEXER_CHECK_MSEC" default:"50ms" max:"200ms"`
	MaxTxAccountCnt             int                `json:"MAX_TX_ACCOUNT_CNT" default:"20" max:"62"`
	FuzzFailPct                 int                `json:"FUZZ_FAIL_PCT" default:"0" max:"0"`
	ConfirmTimeoutSec           time.Duration      `json:"CONFIRM_TIMEOUT_SEC" default:"4s"`
	ConfirmCheckMsec            time.Duration      `json:"CONFIRM_CHECK_MSEC" default:"10ms"`
	MaxEvmStepCntEmulate        int                `json:"MAX_EVM_STEP_CNT_EMULATE" default:"1000" max:"500000"`
	NeonCliTimeout              time.Duration      `json:"NEON_CLI_TIMEOUT" default:"2.5s"`
	NeonCliDebugLog             bool               `json:"NEON_CLI_DEBUG_LOG" default:"false"`
	CancelTimeout               time.Duration      `json:"CANCEL_TIMEOUT" default:"1s" max:"60s"`        // TODO really max timeout can be 1 minute?
	SkipCancelTimeout           time.Duration      `json:"SKIP_CANCEL_TIMEOUT" default:"1s" max:"1000s"` // TODO seconds?
	HolderTimeout               time.Duration      `json:"HOLDER_TIMEOUT" default:"24h" max:"216000h"`   // TODO 9000 days in max?
	GatherStatistics            bool               `json:"GATHER_STATISTICS" default:"false"`
	HvacUrl                     string             `json:"HVAC_URL"`
	HvacToken                   string             `json:"HVAC_TOKEN"`
	HvacMountPoint              string             `json:"HVAC_MOUNT_POINT"`
	GenesisTimestamp            int64              `json:"GENESIS_TIMESTAMP" max:"0"`
	CommitLevel                 rpc.CommitmentType `json:"COMMIT_LEVEL" default:"confirmed"`
	PythMappingAccount          solana.PublicKey   `json:"PYTH_MAPPING_ACCOUNT"`
	UpdatePythMappingPeriodSec  time.Duration      `json:"UPDATE_PYTH_MAPPING_PERIOD_SEC" default:"10s" max:"360s"`
}

func (ic *IndexerConfig) Validate() error {
	switch {
	case ic.MempoolCapacity > mempoolCapacityMax:
		return errors.New("MempoolCapacity exceeds maximum allowed value")
	case ic.MempoolExecutorLimitCnt > mempoolExecutorLimitCnt:
		return errors.New("MempoolExecutorLimitCnt exceeds maximum allowed value")
	case ic.MempoolCacheLifeSec > mempoolCacheLifeSecMax:
		return errors.New("MempoolCacheLifeSec exceeds maximum allowed value")
	case ic.HolderSize > holderSizeMax:
		return errors.New("HolderSize exceeds maximum allowed value")
	case ic.MinOpBalanceToWarn > minOpBalanceToWarnMax:
		return errors.New("MinOpBalanceToWarn exceeds maximum allowed value")
	case ic.MinOpBalanceToErr > minOpBalanceToErrMax:
		return errors.New("MinOpBalanceToErr exceeds maximum allowed value")
	case ic.PermAccountId > permAccountIdMax:
		return errors.New("PermAccountId exceeds maximum allowed value")
	case ic.PermAccountLimit > permAccountLimitMax:
		return errors.New("PermAccountLimit exceeds maximum allowed value")
	case ic.RecheckUsedResourceSec > recheckUsedResourceSecMax:
		return errors.New("RecheckUsedResourceSec exceeds maximum allowed value")
	case ic.RecheckResourceAfterUsesCnt > recheckResourceAfterUsesCntMax:
		return errors.New("RecheckResourceAfterUsesCnt exceeds maximum allowed value")
	case ic.AccountPermissionUpdateInt > accountPermissionUpdateIntMax:
		return errors.New("AccountPermissionUpdateInt exceeds maximum allowed value")
	case ic.SlotProcessingDelay > slotProcessingDelayMax:
		return errors.New("SlotProcessingDelay exceeds maximum allowed value")
	case ic.MinGasPrice.Cmp(decimal.NewFromInt(minGasPriceMax)) > 1:
		return errors.New("MinGasPrice exceeds maximum allowed value")
	case ic.MinWoChainidGasPrice.Cmp(decimal.NewFromInt(minWoChainidGasPriceMax)) > 1:
		return errors.New("MinWoChainidGasPrice exceeds maximum allowed value")
	case ic.NeonDecimals.Cmp(decimal.NewFromInt(neonDecimalsMax)) > 1:
		return errors.New("NeonDecimals exceeds maximum allowed value")
	case ic.IndexerParallelRequestCnt > indexerParallelRequestCntMax:
		return errors.New("IndexerParallelRequestCnt exceeds maximum allowed value")
	case ic.IndexerPollCnt > indexerPollCntMax:
		return errors.New("IndexerPollCnt exceeds maximum allowed value")
	case ic.IndexerLogSkipCnt > indexerLogSkipCntMax:
		return errors.New("IndexerLogSkipCnt exceeds maximum allowed value")
	case ic.IndexerCheckMsec > indexerCheckMsecMax:
		return errors.New("IndexerCheckMsec exceeds maximum allowed value")
	case ic.MaxTxAccountCnt > maxTxAccountCntMax:
		return errors.New("MaxTxAccountCnt exceeds maximum allowed value")
	case ic.FuzzFailPct > fuzzFailPctMax:
		return errors.New("FuzzFailPct exceeds maximum allowed value")
	case ic.ConfirmCheckMsec > confirmCheckMsecMax:
		return errors.New("ConfirmCheckMsec exceeds maximum allowed value")
	case ic.MaxEvmStepCntEmulate > maxEvmStepCntEmulateMax:
		return errors.New("MaxEvmStepCntEmulate exceeds maximum allowed value")
	case ic.GenesisTimestamp > genesisTimestampMax:
		return errors.New("GenesisTimestamp exceeds maximum allowed value")
	case ic.UpdatePythMappingPeriodSec > updatePythMappingPeriodSecMax:
		return errors.New("UpdatePythMappingPeriodSec exceeds maximum allowed value")
	case ic.CommitLevel != rpc.CommitmentConfirmed && ic.CommitLevel != rpc.CommitmentFinalized:
		return errors.New("CommitLevel is lower than 'Confirmed'")
	}
	return nil
}

func (ic *IndexerConfig) setUp() {
	ic.MinGasPrice = ic.MinGasPrice.Mul(decimal.NewFromFloat(math.Pow10(9)))
	ic.MinWoChainidGasPrice = ic.MinWoChainidGasPrice.Mul(decimal.NewFromFloat(math.Pow10(9)))
	ic.IndexerCheckMsec = ic.IndexerCheckMsec / 1000
}

func CreateConfigFromEnv(envPath, fileName string) (cfg *IndexerConfig, err error) {
	cfg = new(IndexerConfig)
	if err := config.LoadConfigFromEnv(cfg, config.WithEnvPath(envPath), config.WithFileName(fileName)); err != nil {
		return nil, err
	}
	return cfg, nil
}
