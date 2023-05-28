package indexer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"strconv"
	"strings"
)

type IxDecoderInterface interface {
	GetIxCode() uint8
	GetName() string
	GetState() SolNeonTxDecoderState
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

// decodingSuccess is called when assembling of the instruction has been finished finished
func (ixd *IxDecoder) decodingSuccess(indxObj any, msg string) bool {
	// The instruction has been successfully parsed
	ixd.log.Debug().Msgf("decoding success: %v -  %v", msg, indxObj)
	return true
}

// decodingSkip is called when assembling of the object can't be finished
func (ixd *IxDecoder) decodingSkip(reason string) bool {
	ixd.log.Warn().Msgf("decoding skip: %v", reason)
	return false
}

// decodingDone is called when assembling of the object has been successfully finished
func (ixd *IxDecoder) decodingDone(indxObj any, msg string) bool {
	block := ixd.state.NeonBlock()

	if neonTxInfo, ok := indxObj.(*NeonIndexedTxInfo); ok {
		block.DoneNeonTx(neonTxInfo, ixd.state.SolNeonIx())
	} else if neonIndexedHolderInfo, ok := indxObj.(*NeonIndexedHolderInfo); ok {
		block.DoneNeonHolder(neonIndexedHolderInfo)
	}

	ixd.log.Debug().Msgf("decoding done: %v - %v", msg, indxObj)
	return true
}

type decodeIterBlockedAccountFunc func() []string

type decodeHolderAccountFunc func() *string

type decodeNeonTxFunc func() *NeonTxInfo

type addReturnEventFunc func(tx *NeonIndexedTxInfo)

type decodeNeonTxReturnFunc func(tx *NeonIndexedTxInfo) bool

// BaseTxIxDecoder implements several methods which are overwritten
// in child classes. To implement such inheritance we use pointers to functions.
// If pointer == nil, BaseTxIxDecoder will invoke his methods instead of child methods implementation
type BaseTxIxDecoder struct {
	*IxDecoder
	decodeHolderAccountFunc
	decodeIterBlockedAccountFunc
	decodeNeonTxFunc
	decodeNeonTxReturnFunc
	addReturnEventFunc
}

func NewBaseTxIxDecoder(ixDecoder *IxDecoder) *BaseTxIxDecoder {
	return &BaseTxIxDecoder{
		IxDecoder: ixDecoder,
	}
}

// addNeonIndexedTx decodes neon tx, holder account and blocked accounts from solana receipt using ixData
// and adds info to block
func (b *BaseTxIxDecoder) addNeonIndexedTx() *NeonIndexedTxInfo {
	neonTx := b.decodeNeonTx()
	if neonTx == nil {
		return nil
	}

	ix := b.state.SolNeonIx()

	// check if decoded neon tx signature from ixData is equal to neon tx signature
	// parsed from solana logs
	if ix.metaInfo.neonTxSig != neonTx.sig {
		b.decodingSkip(fmt.Sprintf("NeonTx.Hash '%s' != SolIx.Log.Hash '%s'", neonTx.sig, ix.metaInfo.neonTxSig))
		return nil
	}

	holderAccount := b.decodeHolderAccount()
	if holderAccount == nil {
		return nil
	}

	iterBlockedAccount := b.decodeIterBlockedAccount()
	if iterBlockedAccount == nil {
		return nil
	}

	// add neon tx to block
	block := b.state.NeonBlock()
	txType := NeonIndexedTxType(b.ixCode)
	return block.AddNeonTx(txType, *neonTx, *holderAccount, iterBlockedAccount, *ix)
}

func (b *BaseTxIxDecoder) decodeNeonTx() *NeonTxInfo {
	if b.decodeNeonTxFunc == nil {
		return NewNeonTxFromNeonSig(b.state.SolNeonIx().metaInfo.neonTxSig)
	}
	// invoke child decodeNeonTx() method implementation
	return b.decodeNeonTxFunc()
}

// decodeHolderAccount decodes holder account
func (b *BaseTxIxDecoder) decodeHolderAccount() *string {
	if b.decodeHolderAccountFunc == nil {
		b.decodeHolderAccountFunc = func() *string {
			panic("Call of not-implemented method to decode NeonHolder.Account")
		}
	}
	// invoke child decodeHolderAccount() method implementation
	return b.decodeHolderAccountFunc()
}

// decodeIterBlockedAccount decodes accounts which are blocked by tx
func (b *BaseTxIxDecoder) decodeIterBlockedAccount() []string {
	if b.decodeIterBlockedAccountFunc == nil {
		b.decodeIterBlockedAccountFunc = func() []string {
			panic("Call of not-implemented method to decode NeonTx.BlockedAccounts")
		}
	}
	// invoke child decodeIterBlockedAccount() method implementation
	return b.decodeIterBlockedAccountFunc()
}

// decodeNeonTxFromHolder decodes neon tx from holder data
func (b *BaseTxIxDecoder) decodeNeonTxFromHolder(holder *NeonIndexedHolderInfo) *NeonTxInfo {
	// decode neon tx from holder data
	neonTx := b.decodeNeonTxFromData("NeonHolder.Data", holder.data)
	if neonTx == nil {
		return nil
	}

	// check if decoded neon tx has same signature with already known tx neon signature for this holder
	if holder.NeonTxSig() != neonTx.sig[2:] {
		b.decodingSkip(fmt.Sprintf("NeonTx.Hash %v != NeonHolder.Hash '%v'", neonTx.sig, holder.NeonTxSig()))
		return nil
	}

	b.decodingDone(holder, fmt.Sprintf("init NeonTx %v from NeonHolder.Data", neonTx))
	return neonTx
}

// decodeNeonTxSigFromIxData decodes neon tx signature from data starting from given offset
func (b *BaseTxIxDecoder) decodeNeonTxSigFromIxData(offset, minLen int) string {
	ix := b.state.SolNeonIx()

	if len(ix.ixData) < minLen {
		b.decodingSkip(fmt.Sprintf("no enough SolIx.Data(len=%v) to get NeonTx.Hash", len(ix.ixData)))
		return ""
	}

	// decode neon tx signature
	neonTxSig := "0x" + strings.ToLower(hex.EncodeToString(ix.ixData[offset:(offset+32)]))

	// check if decoded neon tx signature from ixData is equal to neon tx signature
	// parsed from solana logs
	if ix.metaInfo.neonTxSig != neonTxSig {
		b.decodingSkip(fmt.Sprintf("NeonTx.Hash %v != SolIx.Log.Hash '%v'", neonTxSig, ix.metaInfo.neonTxSig))
		return ""
	}
	return neonTxSig
}

// decodeNeonTxSigFromIxData decodes neon tx from data
func (b *BaseTxIxDecoder) decodeNeonTxFromData(dataName string, data []byte) *NeonTxInfo {
	ix := b.state.SolNeonIx()

	// using rlp decoding get  neon tx from ixData of solana receipt
	neonTx := NewNeonTxFromSigData(data)
	if !neonTx.IsValid() {
		b.decodingSkip(fmt.Sprintf("%v.RLP.Error:'%v'", dataName, neonTx.err))
		return nil

		// check if decoded neon tx signature from ixData is equal to neon tx signature
		// parsed from solana logs
	} else if ix.metaInfo.neonTxSig != neonTx.sig {
		b.decodingSkip(fmt.Sprintf("NeonTx.Hash '%v' != SolIx.Log.Hash '%v'", neonTx.sig, ix.metaInfo.neonTxSig))
		return nil
	}
	return neonTx
}

// decodeNeonTxFromHolderAccount gets holder account from block and if account exists, decode
// neon tx from its data
func (b *BaseTxIxDecoder) decodeNeonTxFromHolderAccount(tx *NeonIndexedTxInfo) bool {
	if tx.neonReceipt.neonTx.IsValid() {
		return false
	}
	ix := b.state.SolNeonIx()
	block := b.state.NeonBlock()

	holder := block.FindNeonTxHolder(tx.storageAccount, *ix)
	if holder == nil {
		return false
	}
	neonTx := b.decodeNeonTxFromHolder(holder)
	if neonTx == nil {
		return false
	}
	tx.SetNeonTx(*neonTx, *holder)
	return true
}

// decodeNeonTxReceipt decodes neon tx events and 'return' neon tx event
func (b *BaseTxIxDecoder) decodeNeonTxReceipt(tx *NeonIndexedTxInfo) bool {
	b.decodeNeonTxEventList(tx)
	if tx.neonReceipt.neonTxRes.IsCompleted() {
		// neon tx is already fully decoded
		return false
	}

	if b.decodeNeonTxReturn(tx) {
		b.addReturnEvent(tx)
		return true
	}
	return false
}

// decodeNeonTxReturn decodes neon tx 'return' event
func (b *BaseTxIxDecoder) decodeNeonTxReturn(tx *NeonIndexedTxInfo) bool {
	if b.decodeNeonTxReturnFunc != nil {
		// invoke child class method implementation
		return b.decodeNeonTxReturnFunc(tx)
	}
	retTx := b.state.SolNeonIx().metaInfo.neonTxReturn
	if retTx == nil {
		return false
	}
	tx.neonReceipt.neonTxRes.SetResult(retTx.Status, retTx.GasUsed)
	return true
}

// decodeNeonTxEventList decodes neon event list
func (b *BaseTxIxDecoder) decodeNeonTxEventList(tx *NeonIndexedTxInfo) {
	solNeonTx := b.state.SolNeonIx()
	totalGasUsed := b.state.solNeonIx.metaInfo.neonTotalGasUsed
	for i := range solNeonTx.metaInfo.neonTxEvents {
		solNeonTx.metaInfo.neonTxEvents[i].totalGasUsed = solNeonTx.metaInfo.neonTotalGasUsed
		solNeonTx.metaInfo.neonTxEvents[i].solSig = solNeonTx.solSign
		solNeonTx.metaInfo.neonTxEvents[i].idx = solNeonTx.metaInfo.idx
		solNeonTx.metaInfo.neonTxEvents[i].innerIdx = solNeonTx.metaInfo.innerIdx

		tx.AddNeonEvent(solNeonTx.metaInfo.neonTxEvents[i])
		totalGasUsed++
	}
}

// addReturnEvent adds decoded 'return' event of neon tx to neon tx
func (b *BaseTxIxDecoder) addReturnEvent(tx *NeonIndexedTxInfo) {
	if b.addReturnEventFunc != nil {
		b.addReturnEventFunc(tx)
		return
	}
	txResult := tx.neonReceipt.neonTxRes

	var eventType LogTxEventType

	if txResult.IsCanceled() {
		eventType = Cancel
	} else if txResult.IsCompleted() {
		eventType = Return
	} else {
		return
	}

	ix := b.state.SolNeonIx()
	txResult.SetSolSigInfo(ix.solSign, ix.metaInfo.idx, ix.metaInfo.innerIdx)

	gasUsed := txResult.gasUsed
	if gasUsed[:2] == "0x" {
		gasUsed = gasUsed[2:]
	}

	totalGasUsed, err := strconv.ParseInt(gasUsed, 16, 64)
	if err != nil {
		b.log.Warn().Msgf("Error parsing totalGasUsed", err)
	}

	event := NeonLogTxEvent{
		eventType:    eventType,
		hidden:       true,
		data:         convertHexStringToLittleEndianByte(txResult.status),
		solSig:       ix.solSign,
		idx:          ix.metaInfo.idx,
		innerIdx:     ix.metaInfo.innerIdx,
		totalGasUsed: totalGasUsed + 5000,
	}
	tx.AddNeonEvent(event)
}

type BaseTxSimpleIxDecoder struct {
	*BaseTxIxDecoder
}

func NewBaseTxSimpleIxDecoder(ixDecoder *IxDecoder) *BaseTxSimpleIxDecoder {
	return &BaseTxSimpleIxDecoder{NewBaseTxIxDecoder(ixDecoder)}
}

func InitBaseTxSimpleIxDecoder(ixDecoder *IxDecoder) *BaseTxSimpleIxDecoder {
	decoder := NewBaseTxSimpleIxDecoder(ixDecoder)
	decoder.BaseTxIxDecoder.decodeIterBlockedAccountFunc = decoder.decodeIterBlockedAccount
	decoder.BaseTxIxDecoder.decodeNeonTxReturnFunc = decoder.decodeNeonTxReturn
	return decoder
}

// decodeTx decodes neon tx
func (b *BaseTxSimpleIxDecoder) decodeTx(msg string) bool {
	tx := b.addNeonIndexedTx()
	if tx == nil {
		return false
	}
	b.decodeNeonTxReceipt(tx)
	b.decodingDone(tx, msg)
	return true
}

func (b *BaseTxSimpleIxDecoder) decodeIterBlockedAccount() []string {
	return []string{}
}

func (b *BaseTxSimpleIxDecoder) decodeNeonTxReturn(tx *NeonIndexedTxInfo) bool {
	// is func pointer = nil,
	// will be called parent method implementation
	b.BaseTxIxDecoder.decodeNeonTxReturnFunc = nil

	// call parent decodeNeonTxReturn() impl
	if b.BaseTxIxDecoder.decodeNeonTxReturn(tx) {
		// set up child method impl for future calls
		b.BaseTxIxDecoder.decodeNeonTxReturnFunc = b.decodeNeonTxReturn
		return true
	}

	// set up child method impl for future calls
	b.BaseTxIxDecoder.decodeNeonTxReturnFunc = b.decodeNeonTxReturn

	ix := b.state.SolNeonIx()
	tx.neonReceipt.neonTxRes.SetLostResult(ix.metaInfo.neonTotalGasUsed)
	b.log.Warn().Msg(fmt.Sprintf("set lost result (is_log_truncated ?= (%v) - %v", ix.metaInfo.isLogTruncated, tx))
	return true
}

// CreateAccount3IxDecoder decodes created neon holder account and adds it to block
type CreateAccount3IxDecoder struct {
	*IxDecoder
}

func (d *CreateAccount3IxDecoder) Execute() bool {
	ix := d.state.SolNeonIx()
	if len(ix.ixData) < 20 {
		return d.decodingSkip(fmt.Sprintf("not enough data to get NeonAccount %d", len(ix.ixData)))
	}

	// get neon account address
	neonAccount := "0x" + hex.EncodeToString(ix.ixData[1:20])
	pdaAccount := ix.GetAccount(2)

	accountInfo := NeonAccountInfo{
		neonAddress: neonAccount,
		pdaAddress:  pdaAccount,
		blockSlot:   ix.blockSlot,
		solSig:      ix.solSign,
	}

	d.state.NeonBlock().AddNeonAccount(accountInfo, *ix)
	return d.decodingSuccess(&accountInfo, "create NeonAccount")
}

type BaseTxStepIxDecoder struct {
	*BaseTxIxDecoder
	firstBlockedAccountIdx int
}

func NewBaseTxStepIxDecoder(ixDecoder *IxDecoder, indx int) *BaseTxStepIxDecoder {
	return &BaseTxStepIxDecoder{
		BaseTxIxDecoder:        NewBaseTxIxDecoder(ixDecoder),
		firstBlockedAccountIdx: indx,
	}
}

func InitBaseTxStepIxDecoder(ixDecoder *IxDecoder, indx int) *BaseTxStepIxDecoder {
	decoder := NewBaseTxStepIxDecoder(ixDecoder, indx)
	decoder.BaseTxIxDecoder.decodeIterBlockedAccountFunc = decoder.decodeIterBlockedAccount
	decoder.BaseTxIxDecoder.decodeHolderAccountFunc = decoder.decodeHolderAccount
	return decoder
}

func (btd *BaseTxStepIxDecoder) decodeTx(msg string) bool {
	if !btd.decodeNeonEvmStepCnt() {
		return false
	}

	tx := btd.getNeonIndexedTx()
	if tx == nil {
		return false
	}

	if btd.decodeNeonTxReceipt(tx) {
		return btd.decodingDone(tx, msg)
	}
	return btd.decodingSuccess(tx, msg)
}

// decodeNeonEvmStepCnt decodes step counter - the number of steps for which the transaction should be executed is known in advance
func (btd *BaseTxStepIxDecoder) decodeNeonEvmStepCnt() bool {
	/*	 1 byte  - ix
		 4 bytes - treasury index
		 4 bytes - neon step cnt
		 4 bytes - unique index */

	ix := btd.state.SolNeonIx()
	if len(ix.ixData) < 9 {
		return btd.decodingSkip(fmt.Sprintf("no enough SolIx.Data(len=%v) to get NeonTx.StepCnt", len(ix.ixData)))
	}
	neonStepCnt := int(binary.LittleEndian.Uint32(ix.ixData[5:9]))
	ix.SetNeonStepCnt(neonStepCnt)
	return true
}

// getNeonIndexedTx checks if some info about neon tx already exists in block
// if not, tries to decode it
func (btd *BaseTxStepIxDecoder) getNeonIndexedTx() *NeonIndexedTxInfo {
	ix := btd.state.SolNeonIx()
	block := btd.state.NeonBlock()
	tx := block.FindNeonTx(*ix)
	if tx == nil {
		return btd.addNeonIndexedTx()
	}
	return tx
}

// decodeHolderAccount gets blocked account by sol tx
func (btd *BaseTxStepIxDecoder) decodeHolderAccount() *string {
	ix := btd.state.SolNeonIx()
	if ix.AccountCnt() < 1 {
		btd.decodingSkip(fmt.Sprintf("no enough SolIx.Accounts(len=%v)", ix.AccountCnt()))
		return nil
	}

	account := ix.GetAccount(0)
	return &account
}

// decodeIterBlockedAccount gets blocked account by sol tx
func (btd *BaseTxStepIxDecoder) decodeIterBlockedAccount() []string {
	ix := btd.state.SolNeonIx()
	if ix.AccountCnt() < btd.firstBlockedAccountIdx+1 {
		btd.decodingSkip(fmt.Sprintf("no enough SolIx.Accounts(len=%v) to get NeonTx.BlockedAccounts", ix.AccountCnt()))
		return nil
	}
	return ix.IterAccount(btd.firstBlockedAccountIdx)
}

func (btd *BaseTxStepIxDecoder) DecodeFailedNeonTxEventList() {
	ix := btd.state.SolNeonIx()

	block := btd.state.NeonBlock()
	tx := block.FindNeonTx(*ix)
	if tx == nil {
		return
	}

	cnt := tx.LenNeonEventList()
	for _, event := range ix.metaInfo.neonTxEvents {
		tx.AddNeonEvent(NeonLogTxEvent{
			eventType:    event.eventType,
			hidden:       true,
			address:      event.address,
			topics:       event.topics,
			data:         event.data,
			solSig:       btd.state.SolNeonIx().solSign,
			idx:          btd.state.SolNeonIx().metaInfo.idx,
			innerIdx:     btd.state.SolNeonIx().metaInfo.innerIdx,
			totalGasUsed: int64(9199999999999999999 + cnt),
			reverted:     true,
			eventLevel:   event.eventLevel,
			eventOrder:   event.eventOrder,
		})
		cnt++

		if ix.metaInfo.isAlreadyFinalized && !tx.neonReceipt.neonTxRes.IsValid() {
			tx.neonReceipt.neonTxRes.SetLostResult(1) // unknown gas usage
			btd.log.Warn().Msg("set lost result")
			btd.decodingDone(tx, "complete by lost result")
		}
	}
}

// TxExecFromDataIxDecoder decodes simple (single) neon tx from sol instruction data
type TxExecFromDataIxDecoder struct {
	*BaseTxSimpleIxDecoder
}

func NewTxExecFromDataIxDecoder(ixDecoder *IxDecoder) *TxExecFromDataIxDecoder {
	return &TxExecFromDataIxDecoder{InitBaseTxSimpleIxDecoder(ixDecoder)}
}

func InitTxExecFromDataIxDecoder(ixDecoder *IxDecoder) *TxExecFromDataIxDecoder {
	decoder := NewTxExecFromDataIxDecoder(ixDecoder)
	decoder.BaseTxIxDecoder.decodeHolderAccountFunc = decoder.decodeHolderAccount
	decoder.BaseTxIxDecoder.decodeNeonTxFunc = decoder.decodeNeonTx
	return decoder
}

func (txd *TxExecFromDataIxDecoder) Execute() bool {
	return txd.decodeTx("NeonTx exec from SolIx.Data")
}

func (txd *TxExecFromDataIxDecoder) decodeHolderAccount() *string {
	addr := ""
	return &addr
}

func (txd *TxExecFromDataIxDecoder) decodeNeonTx() *NeonTxInfo {
	/*	 1 byte  - ix
		 4 bytes - treasury index
		 N bytes - NeonTx */

	ix := txd.state.SolNeonIx()
	if len(ix.ixData) < 6 {
		txd.decodingSkip(fmt.Sprintf("no enough SolIx.Data(len=%v) to decode NeonTx", len(ix.ixData)))
		return nil
	}
	rlpSigData := ix.ixData[5:]
	return txd.decodeNeonTxFromData("SolIx.Data", rlpSigData)
}

// TxExecFromAccountIxDecoder decodes simple (single) neon tx from holder account
type TxExecFromAccountIxDecoder struct {
	*BaseTxSimpleIxDecoder
}

func NewTxExecFromAccountIxDecoder(ixDecoder *IxDecoder) *TxExecFromAccountIxDecoder {
	return &TxExecFromAccountIxDecoder{InitBaseTxSimpleIxDecoder(ixDecoder)}
}

func InitTxExecFromAccountIxDecoder(ixDecoder *IxDecoder) *TxExecFromAccountIxDecoder {
	decoder := NewTxExecFromAccountIxDecoder(ixDecoder)
	decoder.BaseTxIxDecoder.decodeHolderAccountFunc = decoder.decodeHolderAccount
	decoder.BaseTxIxDecoder.addReturnEventFunc = decoder.addReturnEvent
	return decoder
}

func (tad *TxExecFromAccountIxDecoder) Execute() bool {
	return tad.decodeTx("NeonTx exec from NeonHolder.Data")
}

func (tad *TxExecFromAccountIxDecoder) decodeHolderAccount() *string {
	ix := tad.state.SolNeonIx()
	if ix.AccountCnt() < 1 {
		tad.decodingSkip(fmt.Sprintf("no enough SolIx.Accounts(len=%v) to get NeonHolder.Account", ix.AccountCnt()))
		return nil
	}
	account := ix.GetAccount(0)
	return &account
}

func (tad *TxExecFromAccountIxDecoder) addReturnEvent(tx *NeonIndexedTxInfo) {
	// decode neon tx from holder account
	tad.decodeNeonTxFromHolderAccount(tx)

	// call parent implementation of addReturnEvent() method
	tad.BaseTxSimpleIxDecoder.addReturnEventFunc = nil
	tad.BaseTxSimpleIxDecoder.addReturnEvent(tx)
}

// TxStepFromDataIxDecoder decodes iterative neon tx from sol instruction data
type TxStepFromDataIxDecoder struct {
	*BaseTxStepIxDecoder
}

func NewTxStepFromDataIxDecoder(ixDecoder *IxDecoder) *TxStepFromDataIxDecoder {
	return &TxStepFromDataIxDecoder{InitBaseTxStepIxDecoder(ixDecoder, 6)}
}

func InitTxStepFromDataIxDecoder(ixDecoder *IxDecoder) *TxStepFromDataIxDecoder {
	decoder := NewTxStepFromDataIxDecoder(ixDecoder)
	decoder.BaseTxIxDecoder.decodeNeonTxFunc = decoder.decodeNeonTx
	return decoder
}

func (tsd *TxStepFromDataIxDecoder) Execute() bool {
	return tsd.decodeTx("NeonTx step from SolIx.Data")
}

func (tsd *TxStepFromDataIxDecoder) decodeNeonTx() *NeonTxInfo {
	/*	 1 byte  - ix
		 4 bytes - treasury index
		 4 bytes - neon step cnt
		 4 bytes - unique index
		 N bytes - NeonTx */

	ix := tsd.state.SolNeonIx()
	if len(ix.ixData) < 14 {
		tsd.decodingSkip(fmt.Sprintf("no enough SolIx.Data(len=%v) to decode NeonTx", len(ix.ixData)))
		return nil
	}
	rlpSigData := ix.ixData[13:]
	return tsd.decodeNeonTxFromData("SolIx.Data", rlpSigData)
}

// TxStepFromAccountIxDecoder decodes iterative neon tx from holder account
type TxStepFromAccountIxDecoder struct {
	*BaseTxStepIxDecoder
}

func NewTxStepFromAccountIxDecoder(ixDecoder *IxDecoder) *TxStepFromAccountIxDecoder {
	return &TxStepFromAccountIxDecoder{InitBaseTxStepIxDecoder(ixDecoder, 6)}
}

func InitTxStepFromAccountIxDecoder(ixDecoder *IxDecoder) *TxStepFromAccountIxDecoder {
	return NewTxStepFromAccountIxDecoder(ixDecoder)
}

func (tsd *TxStepFromAccountIxDecoder) Execute() bool {
	return tsd.decodeTx("NeonTx step from NeonHolder.Data")
}

func (tsd *TxStepFromAccountIxDecoder) addReturnEvent(tx *NeonIndexedTxInfo) {
	tsd.decodeNeonTxFromHolderAccount(tx)

	tsd.BaseTxStepIxDecoder.addReturnEventFunc = nil
	tsd.BaseTxStepIxDecoder.addReturnEvent(tx)
}

// TxStepFromAccountNoChainIdIxDecoder decodes iterative neon tx, which hasn't chain id, from holder account
type TxStepFromAccountNoChainIdIxDecoder struct {
	*BaseTxStepIxDecoder
}

func NewTxStepFromAccountNoChainIdIxDecoder(ixDecoder *IxDecoder) *TxStepFromAccountNoChainIdIxDecoder {
	return &TxStepFromAccountNoChainIdIxDecoder{InitBaseTxStepIxDecoder(ixDecoder, 6)}
}

func InitTxStepFromAccountNoChainIdIxDecoder(ixDecoder *IxDecoder) *TxStepFromAccountNoChainIdIxDecoder {
	return NewTxStepFromAccountNoChainIdIxDecoder(ixDecoder)
}

func (tsd *TxStepFromAccountNoChainIdIxDecoder) Execute() bool {
	return tsd.decodeTx("NeonTx-wo-ChainId step from NeonHolder.Data")
}

func (tsd *TxStepFromAccountNoChainIdIxDecoder) addReturnEvent(tx *NeonIndexedTxInfo) {
	tsd.decodeNeonTxFromHolderAccount(tx)
	tsd.BaseTxStepIxDecoder.addReturnEvent(tx)
}

type CollectTreasureIxDecoder struct {
	*IxDecoder
}

func (c *CollectTreasureIxDecoder) Execute() bool {
	return c.decodingSuccess(nil, "collect NeonTreasury")
}

// CancelWithHashIxDecoder decodes canceled iterative neon tx
type CancelWithHashIxDecoder struct {
	*BaseTxStepIxDecoder
}

func NewCancelWithHashIxDecoder(ixDecoder *IxDecoder) *CancelWithHashIxDecoder {
	return &CancelWithHashIxDecoder{BaseTxStepIxDecoder: InitBaseTxStepIxDecoder(ixDecoder, 3)}
}

func InitCancelWithHashIxDecoder(ixDecoder *IxDecoder) *CancelWithHashIxDecoder {
	decoder := NewCancelWithHashIxDecoder(ixDecoder)
	decoder.BaseTxIxDecoder.decodeNeonTxReturnFunc = decoder.decodeNeonTxReturn
	return decoder
}

func (chd *CancelWithHashIxDecoder) Execute() bool {
	/* 1  byte  - ix
	   32 bytes - tx hash */

	neonTxSig := chd.decodeNeonTxSigFromIxData(1, 33)
	if len(neonTxSig) == 0 {
		return false
	}

	// get already existing neon tx info or decode it
	tx := chd.getNeonIndexedTx()
	if tx == nil {
		chd.decodingSkip(fmt.Sprintf("cannot find NeonTx '%v'", neonTxSig))
	}
	// decode events list and 'return' event
	chd.decodeNeonTxReceipt(tx)
	return chd.decodingDone(tx, "cancel NeonTx")

}

func (chd *CancelWithHashIxDecoder) decodeNeonTxReturn(tx *NeonIndexedTxInfo) bool {
	tx.neonReceipt.neonTxRes.SetCanceledResult(chd.state.SolNeonIx().metaInfo.neonTotalGasUsed)
	return true
}

type CreateHolderAccountIx struct {
	*IxDecoder
}

func (c *CreateHolderAccountIx) Execute() bool {
	return c.decodingSuccess(nil, "create NeonHolder")
}

type DeleteHolderAccountIx struct {
	*IxDecoder
}

func (d *DeleteHolderAccountIx) Execute() bool {
	return d.decodingSuccess(nil, "delete NeonHolder")
}

// WriteHolderAccountIx decodes neon tx signature, holder account, holder data chunks
// and adds it to block. Decodes neon tx if neon tx info already exists in block, but not fully
type WriteHolderAccountIx struct {
	*BaseTxIxDecoder
}

func NewWriteHolderAccountIx(ixDecoder *IxDecoder) *WriteHolderAccountIx {
	return &WriteHolderAccountIx{NewBaseTxIxDecoder(ixDecoder)}
}

func InitWriteHolderAccountIx(ixDecoder *IxDecoder) *WriteHolderAccountIx {
	return NewWriteHolderAccountIx(ixDecoder)
}

func (w *WriteHolderAccountIx) Execute() bool {
	ix := w.state.SolNeonIx()
	if ix.AccountCnt() < 1 {
		return w.decodingSkip(fmt.Sprintf("no enough SolIx.Accounts(len=%v) to get NeonHolder.Account", ix.AccountCnt()))
	}

	// get holder account key
	holderAccount := ix.GetAccount(0)

	/*	 1  byte  - ix
		 32 bytes - tx hash
		 8  bytes - offset */

	// decode neon tx signature
	neonTxSig := w.decodeNeonTxSigFromIxData(1, 42)
	if len(neonTxSig) == 0 {
		return false
	}

	block := w.state.NeonBlock()

	tx := block.FindNeonTx(*ix)
	if tx != nil && tx.neonReceipt.neonTx.IsValid() {
		// neon tx is fully decoded
		return w.decodingSuccess(tx, "add surplus NeonTx.Data.Chunk to NeonTx")
	}

	holder := block.FindNeonTxHolder(holderAccount, *ix)
	if holder == nil {
		holder = block.AddNeonTxHolder(holderAccount, *ix)
	}

	// get chunk with ethereum (neon) tx
	data := ix.ixData[41:]
	chunk := TxInfoDataChunk{
		offset: int(binary.LittleEndian.Uint32(ix.ixData[33:41])),
		lenght: len(data),
		data:   data,
	}

	holder.AddDataChank(chunk)

	if tx == nil {
		w.decodingSuccess(holder, fmt.Sprintf("add NeonTx.Data.Chunk %v", chunk))
		return true
	}

	// decode neon tx from holder data chunk
	neonTx := w.decodeNeonTxFromHolder(holder)
	if neonTx != nil {
		tx.SetNeonTx(*neonTx, *holder)
	}
	w.decodingSuccess(holder, fmt.Sprintf("add NeonTx.Data.Chunk %v", chunk))
	return true
}

type Deposit3IxDecoder struct {
	*IxDecoder
}

func (d *Deposit3IxDecoder) Execute() bool {
	return d.decodingSuccess(nil, "deposit NEONs")
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
		InitTxExecFromDataIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionExecuteFromInstruction",
			ixCode: 0x1f,
		}),
		InitTxExecFromAccountIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionExecFromAccount",
			ixCode: 0x2a,
		}),
		InitTxStepFromDataIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromInstruction",
			ixCode: 0x20,
		}),
		InitTxStepFromAccountIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromAccount",
			ixCode: 0x21,
		}),
		InitTxStepFromAccountNoChainIdIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromAccountNoChainId",
			ixCode: 0x22,
		}),
		InitCancelWithHashIxDecoder(&IxDecoder{
			log:    log,
			name:   "CancelWithHash",
			ixCode: 0x23,
		}),
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
		InitWriteHolderAccountIx(&IxDecoder{
			log:    log,
			name:   "WriteHolderAccount",
			ixCode: 0x26,
		}),
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
	if len(hexString) > 2 {
		if hexString[:2] == "0x" {
			hexString = hexString[2:] //  skip 0x
		}
	}

	hexInt, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		panic("error converting string to hex number")
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(hexInt))
	return buf[:1]
}
