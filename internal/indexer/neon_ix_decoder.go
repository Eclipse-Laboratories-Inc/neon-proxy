package indexer

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"strconv"
	"strings"
)

type IxDecoderInterface interface {
	GetIxCode() uint8
	GetName() string
	GetState() SolNeonTxDecoderState
	InitState(SolNeonTxDecoderState)

	IsDeprecated() bool
	Execute() bool
}

type IxDecoder struct {
	log          logger.Logger
	name         string
	ixCode       uint8
	state        SolNeonTxDecoderState
	isDeprecated bool
}

func NewIxDecoder(log logger.Logger, name string, ixCode uint8, deprecated bool, state SolNeonTxDecoderState) *IxDecoder {
	return &IxDecoder{
		log:          log,
		name:         name,
		ixCode:       ixCode,
		state:        state,
		isDeprecated: deprecated,
	}
}

func (ixd *IxDecoder) GetIxCode() uint8 {
	return ixd.ixCode
}

func (ixd *IxDecoder) IsDeprecated() bool {
	return ixd.isDeprecated
}

func (ixd *IxDecoder) GetName() string {
	return ixd.name
}

func (ixd *IxDecoder) Execute() bool {
	ix := ixd.state.SolNeonIx()
	return ixd.decodingSkip(fmt.Sprintf("no logic to decode the instruction %v", hex.EncodeToString(ix.ixData[:8])))
}

func (ixd *IxDecoder) GetState() SolNeonTxDecoderState {
	return ixd.state
}

func (ixd *IxDecoder) InitState(state SolNeonTxDecoderState) {
	ixd.state = state
}

// TODO check
func (ixd *IxDecoder) decodingSuccess(indxObj any, msg string) bool {
	ixd.log.Debug().Msgf("decoding success: %v -  %v", msg, indxObj)
	return true
}

// TODO check
func (ixd *IxDecoder) decodingSkip(reason string) bool {
	ixd.log.Warn().Msgf("decoding skip: %v", reason)
	return true
}

func (ixd *IxDecoder) decodingDone(indxObj any, msg string) bool {
	ix := ixd.state.SolNeonIx()
	block := ixd.state.NeonBlock()

	if neonTxInfo, ok := indxObj.(NeonIndexedTxInfo); ok {
		block.DoneNeonTx(neonTxInfo, ix)
	} else if neonIndexedHolderInfo, ok := indxObj.(NeonIndexedHolderInfo); ok {
		block.DoneNeonHolder(neonIndexedHolderInfo)
	} else {
		ixd.log.Warn().Msgf("unknown indexed object type: %v - %v", msg, indxObj)
		return true
	}

	ixd.log.Debug().Msgf("decoding done: %v - %v", msg, indxObj)
	return true
}

func (ixd *IxDecoder) decodeNeonTxFromHolder(tx *NeonIndexedTxInfo) {
	// TODO check (видимо потому что если адрес не пустой, декодеровать нечего (?)
	if tx.neonReceipt.neonTx.IsValid() || !tx.neonReceipt.neonTxRes.IsValid() {
		return
	}

	if tx.txType != NeonIndexedTxTypeSingleFromAccount &&
		tx.txType != NeonIndexedTxTypeIterFromAccount &&
		tx.txType != NeonIndexedTxTypeIterFromAccountWoChainId {
		return
	}

	key := TxInfoKey{neonTxSig: tx.neonReceipt.neonTx.sig, account: tx.storageAccount}
	holder := ixd.state.NeonBlock().FindNeonTxHolder(key, *ixd.state.SolNeonIx())
	if holder == nil {
		return
	}

	neonTx := NewNeonTxFromSigData(holder.data)
	if !neonTx.IsValid() {
		ixd.decodingSkip(fmt.Sprintf("Neon tx rlp error: %s", neonTx.err))
	} else if holder.NeonTxSig() != neonTx.sig[2:] {
		ixd.decodingSkip(fmt.Sprintf("Neon tx hash %s != holder hash %s", neonTx.sig, holder.NeonTxSig()))
	} else if neonTx.sig != tx.neonReceipt.neonTx.sig {
		ixd.decodingSkip(fmt.Sprintf("Neon tx hash %s != tx log hash %s", neonTx.sig, tx.neonReceipt.neonTx.sig))
	} else {
		tx.SetNeonTx(*neonTx, *holder)
		ixd.decodingDone(holder, fmt.Sprintf("init Neon tx %s from holder", tx.neonReceipt.neonTx))
	}
}

func (ixd *IxDecoder) decodeNeonTxReturn(tx *NeonIndexedTxInfo) {
	neonTxResult := tx.neonReceipt.neonTxRes
	if neonTxResult.IsValid() {
		return
	}

	ix := ixd.state.SolNeonIx()
	neonTxReturn := ix.metaInfo.neonTxReturn
	if neonTxReturn != nil {
		neonTxResult.SetResult(neonTxReturn.Status, neonTxReturn.GasUsed)
		// TODO check
	} else if !ixd.decodeNeonTxCancelReturn(tx) && !ixd.decodeNeonTxLostReturn(tx) {
		return
	}

	neonTxResult.SetSolSigInfo(ix.solSign, ix.metaInfo.idx, ix.metaInfo.innerIdx)

	eventType := Cancel
	if !tx.IsCanceled() {
		eventType = Return
	}

	gasUsed, err := strconv.ParseInt(neonTxResult.gasUsed[2:], 16, 64)
	if err != nil {
		ixd.log.Error().Err(err).Msg("error converting 'gas used' hex string to number")
	}

	tx.AddNeonEvent(NeonLogTxEvent{
		eventType:    eventType,
		Hidden:       true,
		topics:       []string{},
		data:         convertHexStringToLittleEndianByte(neonTxReturn.Status),
		totalGasUsed: gasUsed + 5000, // to move event to the end of the list
		solSig:       ix.solSign,
		idx:          ix.metaInfo.idx,
		innerIdx:     ix.metaInfo.innerIdx,
	})
}

func (ixd *IxDecoder) decodeNeonTxCancelReturn(tx *NeonIndexedTxInfo) bool {
	if tx.txType == NeonIndexedTxTypeSingleFromAccount ||
		tx.txType == NeonIndexedTxTypeSingle {
		return false
	}

	ix := ixd.state.solNeonIx
	if !tx.IsCanceled() {
		return false
	}

	tx.neonReceipt.neonTxRes.SetCanceledResult(ix.metaInfo.neonTotalGasUsed)
	return true
}

// TODO does nothing may be remove?
func (ixd *IxDecoder) decodeNeonTxLostReturn(tx *NeonIndexedTxInfo) bool {
	return false
}

func (ixd *IxDecoder) decodeNeonTxEventList(tx *NeonIndexedTxInfo) {
	totalGasUsed := ixd.state.solNeonIx.metaInfo.neonTotalGasUsed
	solNeonTx := ixd.state.SolNeonIx()
	for _, event := range solNeonTx.metaInfo.neonTxEvents {
		tx.AddNeonEvent(NeonLogTxEvent{
			eventType:    event.eventType,
			Hidden:       event.Hidden,
			address:      event.address,
			topics:       event.topics,
			data:         event.data,
			totalGasUsed: totalGasUsed,
			solSig:       solNeonTx.solSign,
			idx:          solNeonTx.metaInfo.idx,
			innerIdx:     solNeonTx.metaInfo.innerIdx,
		})
		totalGasUsed++
	}
}

func (ixd *IxDecoder) decodeTx(tx *NeonIndexedTxInfo, msg string) bool {
	ixd.decodeNeonTxReturn(tx)
	ixd.decodeNeonTxEventList(tx)
	ixd.decodeNeonTxFromHolder(tx)

	if tx.neonReceipt.neonTxRes.IsValid() && tx.status != NeonIndexedTxInfoStatusDone {
		return ixd.decodingDone(tx, msg)
	}

	return ixd.decodingSuccess(tx, msg)
}

type CreateAccount3IxDecoder struct {
	*IxDecoder
}

func (d *CreateAccount3IxDecoder) Execute() bool {
	ix := d.state.SolNeonIx()
	if len(ix.ixData) < 20 {
		return d.decodingSkip(fmt.Sprintf("not enough data to get Neon account %d", len(ix.ixData)))
	}

	neonAccount := "0x" + hex.EncodeToString(ix.ixData[1:21])
	pdaAccount := ix.GetAccount(2)

	accountInfo := NeonAccountInfo{
		neonAddress: neonAccount,
		pdaAddress:  pdaAccount,
		blockSlot:   ix.blockSlot,
		solSig:      ix.solSign,
	}

	d.state.NeonBlock().AddNeonAccount(accountInfo, *ix)
	return d.decodingSuccess(&accountInfo, "create Neon account")
}

type BaseTxStepIxDecoder struct {
	*IxDecoder
	txType                 NeonIndexedTxType
	firstBlockedAccountIdx int
}

func (btd *BaseTxStepIxDecoder) getNeonTx() *NeonIndexedTxInfo {
	ix := btd.state.SolNeonIx()
	if ix == nil {
		btd.log.Error().Err(errors.New("no solana Neon tx"))
		return nil
	}

	if ix.AccountCnt() < btd.firstBlockedAccountIdx+1 {
		btd.decodingSkip("no enough accounts")
		return nil
	}

	if len(ix.ixData) < 9 {
		btd.decodingSkip("no enough data to get Neon step cnt")
		return nil
	}

	neonStepCnt := int(binary.LittleEndian.Uint32(ix.ixData[5:9]))
	ix.SetNeonStepCnt(neonStepCnt)

	neonTxSig := ix.metaInfo.neonTxSig
	if len(neonTxSig) == 0 {
		btd.decodingSkip("no Neon tx hash in logs")
		return nil
	}

	block := btd.state.NeonBlock()
	key := NewTxInfoKey(*ix)
	tx := block.FindNeonTx(key, *ix)
	if tx != nil {
		return tx
	}

	neonTx := NewNeonTxFromNeonSig(neonTxSig)
	holderAccount := btd.getHolderAccount()
	iterBlockedAccount := btd.iterBlockedAccount()
	return block.AddNeonTx(btd.txType, key, *neonTx, holderAccount, iterBlockedAccount, *ix)
}

func (btd *BaseTxStepIxDecoder) getHolderAccount() string {
	return btd.state.SolNeonIx().GetAccount(0)
}

func (btd *BaseTxStepIxDecoder) iterBlockedAccount() []string {
	return btd.state.SolNeonIx().IterAccount(btd.firstBlockedAccountIdx)
}

// TODO remove (?)
func (btd *BaseTxStepIxDecoder) decodeNeonTxLostReturn() bool {
	return false
}

func (btd *BaseTxStepIxDecoder) DecodeFailedNeonTxEventList() {
	ix := btd.state.SolNeonIx()
	if ix == nil {
		btd.log.Error().Err(errors.New("no solana Neon tx"))
		return
	}
	block := btd.state.NeonBlock()
	key := NewTxInfoKey(*ix)
	tx := block.FindNeonTx(key, *ix)
	if tx == nil {
		return
	}

	for _, event := range ix.metaInfo.neonTxEvents {
		tx.AddNeonEvent(NeonLogTxEvent{
			eventType:    event.eventType,
			Hidden:       true,
			reverted:     true,
			address:      event.address,
			topics:       event.topics,
			data:         event.data,
			totalGasUsed: int64(tx.LenNeonEventList()),
			solSig:       btd.state.SolNeonIx().solSign,
			idx:          btd.state.SolNeonIx().metaInfo.idx,
			innerIdx:     btd.state.SolNeonIx().metaInfo.innerIdx,
		})
	}
}

type TxExecFromDataIxDecoder struct {
	*BaseTxStepIxDecoder
}

func (txd *TxExecFromDataIxDecoder) Execute() bool {
	ix := txd.state.SolNeonIx()
	if len(ix.ixData) < 6 {
		return txd.decodingSkip("no enough data to get Neon tx")
	}

	rlpSigData := ix.ixData[5:]
	neonTx := NewNeonTxFromSigData(rlpSigData)
	if neonTx.err != nil {
		return txd.decodingSkip(fmt.Sprintf("Neon tx rlp error %v", neonTx.err))
	}

	if neonTx.sig != ix.metaInfo.neonTxSig {
		return txd.decodingSkip(fmt.Sprintf("Neon tx hash %v != hash %v", neonTx.sig, ix.metaInfo.neonTxSig))
	}

	tx := txd.getNeonTx()
	if tx == nil {
		return false
	}
	tx.SetNeonTx(*neonTx, NeonIndexedHolderInfo{})
	return txd.decodeTx(tx, "Neon tx exec from data")
}

type TxExecFromAccountIxDecoder struct {
	*BaseTxStepIxDecoder
}

func (tad *TxExecFromAccountIxDecoder) Execute() bool {
	ix := tad.state.SolNeonIx()
	if len(ix.ixData) < 6 {
		return tad.decodingSkip("no enough data for ix data")
	}

	if ix.AccountCnt() < 1 {
		return tad.decodingSkip("no enough accounts to get holder account")
	}

	tx := tad.getNeonTx()
	if tx == nil {
		return false
	}
	return tad.decodeTx(tx, "Neon tx exec from account")
}

type TxStepFromDataIxDecoder struct {
	*BaseTxStepIxDecoder
}

func (tsd *TxStepFromDataIxDecoder) Execute() bool {
	tx := tsd.getNeonTx()
	if tx == nil {
		return false
	}

	if tx.neonReceipt.neonTx.IsValid() {
		tsd.decodeTx(tx, "Neon tx continue step from data")
	}

	ix := tsd.state.SolNeonIx()
	if len(ix.ixData) < 14 {
		return tsd.decodingSkip("no enough data to get Neon tx")
	}

	rlpSigData := ix.ixData[13:]
	neonTx := NewNeonTxFromSigData(rlpSigData)
	if neonTx.err != nil {
		return tsd.decodingSkip(fmt.Sprintf("Neon tx rlp error %v", neonTx.err))
	}
	if neonTx.sig != tx.neonReceipt.neonTx.sig {
		return tsd.decodingSkip(fmt.Sprintf("Neon tx hash %v != tx log hash %v", neonTx.sig, tx.neonReceipt.neonTx.sig))
	}
	tx.SetNeonTx(*neonTx, NeonIndexedHolderInfo{})
	return tsd.decodeTx(tx, "Neon tx init step from data")
}

type TxStepFromAccountIxDecoder struct {
	*BaseTxStepIxDecoder
}

func (tsd *TxStepFromAccountIxDecoder) Execute() bool {
	tx := tsd.getNeonTx()
	if tx == nil {
		return false
	}
	return tsd.decodeTx(tx, "Neon tx step from account")
}

type TxStepFromAccountNoChainIdIxDecoder struct {
	*BaseTxStepIxDecoder
}

func (tsd *TxStepFromAccountNoChainIdIxDecoder) Execute() bool {
	tx := tsd.getNeonTx()
	if tx == nil {
		return false
	}
	return tsd.decodeTx(tx, "Neon tx wo chain-id step from account")
}

type CollectTreasureIxDecoder struct {
	*IxDecoder
}

type CancelWithHashIxDecoder struct {
	*IxDecoder
	firstBlockedAccountIdx int
}

func (chd *CancelWithHashIxDecoder) Execute() bool {
	ix := chd.state.SolNeonIx()
	if ix.AccountCnt() < chd.firstBlockedAccountIdx+1 {
		return chd.decodingSkip("no enough accounts")
	}

	if len(ix.ixData) < 33 {
		return chd.decodingSkip(fmt.Sprintf("no enough data to get Neon tx hash: hash len %v", len(ix.ixData)))
	}
	neonTxSig := "0x" + strings.ToLower(hex.EncodeToString(ix.ixData[1:33]))
	logTxSig := ix.metaInfo.neonTxSig

	if logTxSig != neonTxSig {
		return chd.decodingSkip(fmt.Sprintf("Neon tx hash %v != %v", logTxSig, neonTxSig))
	}

	key := NewTxInfoKey(*ix)
	tx := chd.state.NeonBlock().FindNeonTx(key, *ix)
	if tx == nil {
		return chd.decodingSkip(fmt.Sprintf("cannot find Neon tx %v", neonTxSig))
	}
	tx.canceled = true
	return chd.decodeTx(tx, "cancel Neon tx")
}

type CreateHolderAccountIx struct {
	*IxDecoder
}

type DeleteHolderAccountIx struct {
	*IxDecoder
}

type WriteHolderAccountIx struct {
	*IxDecoder
}

func (w *WriteHolderAccountIx) Execute() bool {
	ix := w.state.SolNeonIx()
	if ix.AccountCnt() < 1 {
		return w.decodingSkip(fmt.Sprintf("no enough accounts: %v", ix.AccountCnt()))
	}

	if len(ix.ixData) < 42 {
		return w.decodingSkip(fmt.Sprintf("no enough data to get Neon tx data chunk: chunk len %v", len(ix.ixData)))
	}

	data := ix.ixData[41:]
	chunk := TxInfoDataChunk{
		offset: int(binary.LittleEndian.Uint32(ix.ixData[33:41])),
		lenght: len(data),
		data:   data,
	}

	neonTxSig := "0x" + strings.ToLower(hex.EncodeToString(ix.ixData[1:33]))
	logTxSig := ix.metaInfo.neonTxSig

	if logTxSig != neonTxSig {
		return w.decodingSkip(fmt.Sprintf("Neon tx hash %v != %v", logTxSig, neonTxSig))
	}

	block := w.state.NeonBlock()
	holderAccount := ix.GetAccount(0)

	key := NewTxInfoKey(*ix)
	tx := block.FindNeonTx(key, *ix)
	if tx != nil && tx.neonReceipt.neonTx.IsValid() {
		return w.decodingSuccess(tx, "add surplus data chunk to tx")
	}

	key = TxInfoKey{
		neonTxSig: neonTxSig,
		account:   holderAccount,
	}

	holder := block.FindNeonTxHolder(key, *ix)
	if holder == nil {
		holder = block.AddNeonTxHolder(key, *ix)
	}
	holder.AddDataChank(chunk)
	w.decodingSuccess(holder, fmt.Sprintf("add Neon tx data chunk %v", chunk))
	if tx != nil {
		w.decodeNeonTxFromHolder(tx)
	}
	return true
}

type Deposit3IxDecoder struct {
	*IxDecoder
}

func GetNeonIxDecoderList(log logger.Logger) []IxDecoderInterface {
	ixDecoderList := []IxDecoderInterface{
		&CreateAccount3IxDecoder{
			&IxDecoder{
				log:    log,
				name:   "CreateAccount3",
				ixCode: 0x28,
			},
		},
		&CollectTreasureIxDecoder{
			&IxDecoder{
				log:    log,
				name:   "CollectTreasure",
				ixCode: 0x1e,
			},
		},
		&TxExecFromDataIxDecoder{
			&BaseTxStepIxDecoder{
				IxDecoder: &IxDecoder{
					log:    log,
					name:   "TransactionExecuteFromInstruction",
					ixCode: 0x1f,
				},
				txType:                 NeonIndexedTxTypeSingle,
				firstBlockedAccountIdx: 6,
			},
		},
		&TxExecFromAccountIxDecoder{
			&BaseTxStepIxDecoder{
				IxDecoder: &IxDecoder{
					log:    log,
					name:   "TransactionExecFromAccount",
					ixCode: 0x2a,
				},
				txType:                 NeonIndexedTxTypeSingleFromAccount,
				firstBlockedAccountIdx: 6,
			},
		},
		&TxStepFromDataIxDecoder{
			&BaseTxStepIxDecoder{
				IxDecoder: &IxDecoder{
					log:    log,
					name:   "TransactionStepFromInstruction",
					ixCode: 0x20,
				},
				txType:                 NeonIndexedTxTypeIterFromData,
				firstBlockedAccountIdx: 6,
			},
		},
		&TxStepFromAccountIxDecoder{
			&BaseTxStepIxDecoder{
				IxDecoder: &IxDecoder{
					log:    log,
					name:   "TransactionStepFromAccount",
					ixCode: 0x21,
				},
				txType:                 NeonIndexedTxTypeIterFromAccount,
				firstBlockedAccountIdx: 6,
			},
		},
		&TxStepFromAccountNoChainIdIxDecoder{
			&BaseTxStepIxDecoder{
				IxDecoder: &IxDecoder{
					log:    log,
					name:   "TransactionStepFromAccountNoChainId",
					ixCode: 0x22,
				},
				txType:                 NeonIndexedTxTypeIterFromAccountWoChainId,
				firstBlockedAccountIdx: 6,
			},
		},
		&CancelWithHashIxDecoder{
			IxDecoder: &IxDecoder{
				log:    log,
				name:   "CancelWithHash",
				ixCode: 0x23,
			},
			firstBlockedAccountIdx: 3,
		},
		&CreateHolderAccountIx{
			IxDecoder: &IxDecoder{
				log:    log,
				name:   "CreateHolderAccount",
				ixCode: 0x24,
			},
		},
		&DeleteHolderAccountIx{
			IxDecoder: &IxDecoder{
				log:    log,
				name:   "DeleteHolderAccount",
				ixCode: 0x25,
			},
		},
		&WriteHolderAccountIx{
			IxDecoder: &IxDecoder{
				log:    log,
				name:   "WriteHolderAccount",
				ixCode: 0x26,
			},
		},
		&Deposit3IxDecoder{
			IxDecoder: &IxDecoder{
				log:    log,
				name:   "Deposit3",
				ixCode: 0x27,
			},
		},
	}

	for _, ixDecoder := range ixDecoderList {
		if ixDecoder.IsDeprecated() {
			panic(fmt.Sprintf("%s is deprecated!", ixDecoder.GetName()))
		}
	}
	return ixDecoderList
}

func convertHexStringToLittleEndianByte(hexString string) []byte {
	hexBytes := hexString[2:] //  skip 0x
	hexInt := uint64(0)
	fmt.Sscanf(hexBytes, "%x", &hexInt)
	buf := make([]byte, 1)
	binary.LittleEndian.PutUint16(buf, uint16(hexInt))
	return buf
}
