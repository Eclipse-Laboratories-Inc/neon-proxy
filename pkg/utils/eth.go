package utils

import (
	"bytes"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

// For mode info about rlp format
// https://ethereum.org/es/developers/docs/data-structures-and-encoding/rlp/

type NeonTx struct {
	Nonce     *big.Int
	GasPrice  *big.Int
	GasLimit  *big.Int
	ToAddress []byte
	Value     *big.Int
	CallData  []byte
	V         *big.Int
	R         *big.Int
	S         *big.Int
}

type NeonNoChainTx struct {
	Nonce     *big.Int
	GasPrice  *big.Int
	GasLimit  *big.Int
	ToAddress []byte
	Value     *big.Int
	CallData  []byte
}

func NewNeonNoChainTxFromString(rlpData []byte) (*NeonNoChainTx, error) {
	var tx NeonNoChainTx
	err := rlp.DecodeBytes(rlpData, &tx)
	return &tx, err
}

func NewNeonTxFromString(rlpData []byte) (*NeonTx, error) {
	var tx NeonTx
	if err := rlp.DecodeBytes(rlpData, &tx); err != nil {
		noChainTx, err := NewNeonNoChainTxFromString(rlpData)
		if err != nil {
			return nil, err
		}
		return &NeonTx{
			Nonce:     noChainTx.Nonce,
			GasPrice:  noChainTx.GasPrice,
			GasLimit:  noChainTx.GasLimit,
			ToAddress: noChainTx.ToAddress,
			Value:     noChainTx.Value,
			CallData:  noChainTx.CallData,
		}, nil
	}
	return &tx, nil
}

func (tx *NeonTx) TxSig() []byte {
	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic("cant' encode tx")
	}
	return crypto.Keccak256(encodedTx)
}

func (tx *NeonTx) HexTxSig() string {
	sig := tx.TxSig()
	return "0x" + hex.EncodeToString(sig)
}

// Sender gets address 'from' of Ethereum transaction using recovered public key of transaction sender.
// It recovers public key recovering tx signature from 'v', 'r' and 's' values and tx hash
func (tx *NeonTx) Sender() []byte {
	secpk1, _ := new(big.Int).SetString("115792089237316195423570985008687907852837564279074904382605163141518161494337", 10)
	nullAddress := bytes.Repeat([]byte{0xff}, 20)

	if tx.R.Cmp(big.NewInt(0)) == 0 && tx.S.Cmp(big.NewInt(0)) == 0 {
		//  tx is unsigned
		return nullAddress
	} else if !tx.HasChainID() {
		// this condition made to go forward if tx doesn't have chain id
	} else if tx.V.Cmp(big.NewInt(37)) >= 0 {
		// get value of chain id from 'v'
		vee := new(big.Int).Sub(tx.V, new(big.Int).Mul(tx.ChainID(), big.NewInt(2)))
		vee.Sub(vee, big.NewInt(8))
		if vee.Cmp(big.NewInt(27)) != 0 && vee.Cmp(big.NewInt(28)) != 0 {
			panic("Invalid V value")
		}
	} else {
		panic("Invalid V value")
	}

	// verify signature values
	if tx.R.Cmp(secpk1) >= 0 || tx.S.Cmp(secpk1) >= 0 ||
		tx.R.Cmp(big.NewInt(0)) == 0 || tx.S.Cmp(big.NewInt(0)) == 0 {
		panic("Invalid signature values")
	}

	sigHash, err := tx.unsignedMsgImpl()
	if err != nil {
		panic("Failed to get hash")
	}
	sigHash = crypto.Keccak256(sigHash)

	// recover tx signature
	sig := tx.sigImpl()

	// recover public key
	pubKeyBytes, err := crypto.Ecrecover(sigHash, sig)
	if err != nil {
		panic("Failed to recover public key")
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		panic("Failed to unmarshal public key")
	}

	return crypto.PubkeyToAddress(*pubKey).Bytes()
}

// sigImpl recovers signature from 'v', 'r', 's' fields
func (tx *NeonTx) sigImpl() []byte {
	sig := make([]byte, 0)

	sig = append(sig, tx.R.Bytes()...)
	sig = append(sig, tx.S.Bytes()...)
	if tx.V.Int64() == 0 {
		sig = append(sig, 0)
	} else {
		sig = append(sig, 1)
	}
	return sig
}

// HexSender get address of Ethereum transaction sender as hex string
func (tx *NeonTx) HexSender() string {
	return "0x" + hex.EncodeToString(tx.Sender())
}

func (tx *NeonTx) HasChainID() bool {
	// if v = 27 or v = 28, it shows the way how to recover
	// public key from signature (used in older Ethereum versions, but still valid)
	return tx.V.Int64() != 0 && tx.V.Int64() != 27 && tx.V.Int64() != 28
}

func (tx *NeonTx) ChainID() *big.Int {
	if !tx.HasChainID() {
		return nil
	}
	if tx.V.Int64() >= 37 {
		// chainid*2 + 35  xxxxx0 + 100011   xxxx0 + 100010 +1
		// chainid*2 + 36  xxxxx0 + 100100   xxxx0 + 100011 +1
		return big.NewInt((tx.V.Int64()-1)/2 - 17)
	}
	panic("Invalid V value")
}

func (tx *NeonTx) unsignedMsgImpl() ([]byte, error) {
	if !tx.HasChainID() {
		return rlp.EncodeToBytes(&NeonNoChainTx{
			Nonce:     tx.Nonce,
			GasPrice:  tx.GasPrice,
			GasLimit:  tx.GasLimit,
			ToAddress: tx.ToAddress,
			Value:     tx.Value,
			CallData:  tx.CallData,
		})
	} else {
		return rlp.EncodeToBytes(&NeonTx{
			Nonce:     tx.Nonce,
			GasPrice:  tx.GasPrice,
			GasLimit:  tx.GasLimit,
			ToAddress: tx.ToAddress,
			Value:     tx.Value,
			CallData:  tx.CallData,
			V:         tx.ChainID(),  // chain id, used to avoid signature conflicts between different Ethereum chains (testnet, mainnet and etc.)
			R:         big.NewInt(0), // is 'r' and 's' fields = 0, tx is unsigned
			S:         big.NewInt(0),
		})
	}
}

func (tx *NeonTx) HexCallData() string {
	return "0x" + hex.EncodeToString(tx.CallData)
}

func (tx *NeonTx) Contract() []byte {
	if tx.ToAddress != nil {
		return nil
	}
	type contractData struct {
		sender []byte
		nonce  *big.Int
	}

	contractAddr, err := rlp.EncodeToBytes(&contractData{sender: tx.Sender(), nonce: tx.Nonce})
	if err != nil {
		panic("can't encode contract data")
	}
	contract := crypto.Keccak256(contractAddr)
	return contract[len(contract)-20:]
}

func (tx *NeonTx) HexContract() string {
	contract := tx.Contract()
	if contract == nil {
		return ""
	}
	return "0x" + hex.EncodeToString(contract)
}

func (tx *NeonTx) HexToAddress() string {
	if tx.ToAddress == nil {
		return ""
	}

	return "0x" + hex.EncodeToString(tx.ToAddress)
}
