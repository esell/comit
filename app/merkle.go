package app

import (
	"fmt"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmsp "github.com/tendermint/tmsp/types"
)

const (
	ErrValueNotFound = 311311
)

type MerkleApp struct {
	tree *merkle.IAVLTree
}

func NewMerkleApp() *MerkleApp {
	tree := merkle.NewIAVLTree(0, nil)
	return &MerkleApp{tree}
}

func (merk *MerkleApp) Info() string {
	return Fmt("size:%v", merk.tree.Size())
}

func (merk *MerkleApp) SetOption(key string, value string) (log string) {
	return "No options are supported yet"
}

func (merk *MerkleApp) AppendTx(txBytes []byte) tmsp.Result {
	if len(txBytes) == 0 {
		return tmsp.ErrEncodingError.SetLog(
			"Tx length must be greater than zero")
	}
	typeByte := txBytes[0]
	txBytes = txBytes[1:]
	key, n, err := wire.GetByteSlice(txBytes)
	if err != nil {
		return tmsp.ErrEncodingError.SetLog(
			Fmt("Error getting key: %v", err.Error()))
	}
	txBytes = txBytes[n:]
	switch typeByte {
	case 0x01:
		value, n, err := wire.GetByteSlice(txBytes)
		if err != nil {
			return tmsp.ErrEncodingError.SetLog(
				Fmt("Error getting value: %v", err.Error()))
		}
		txBytes = txBytes[n:]
		if len(txBytes) != 0 {
			return tmsp.ErrEncodingError.SetLog("Got bytes left over")
		}
		merk.tree.Set(key, value)
		fmt.Println("SET", Fmt("%X", key), Fmt("%X", value))
	case 0x02:
		if len(txBytes) != 0 {
			return tmsp.ErrEncodingError.SetLog("Got bytes left over")
		}
		merk.tree.Remove(key)
	default:
		return tmsp.ErrUnknownRequest.SetLog(
			Fmt("Unexpected AppendTx type byte %X", typeByte))
	}
	return tmsp.OK
}

func (merk *MerkleApp) CheckTx(txBytes []byte) tmsp.Result {
	if len(txBytes) == 0 {
		return tmsp.ErrEncodingError.SetLog(
			"Tx length must be greater than zero")
	}
	_, n, err := wire.GetByteSlice(txBytes)
	if err != nil {
		return tmsp.ErrEncodingError.SetLog(
			Fmt("Error getting key: %v", err.Error()))
	}
	txBytes = txBytes[n:]
	_, n, err = wire.GetByteSlice(txBytes)
	if err != nil {
		return tmsp.ErrEncodingError.SetLog(
			Fmt("Error getting value: %v", err.Error()))
	}
	txBytes = txBytes[n:]
	if len(txBytes) != 0 {
		return tmsp.ErrEncodingError.SetLog("Got bytes left over")
	}
	return tmsp.OK
}

func (merk *MerkleApp) Commit() tmsp.Result {
	if merk.tree.Size() == 0 {
		return tmsp.NewResultOK(nil, "Empty hash for empty tree")
	}
	hash := merk.tree.Hash()
	return tmsp.NewResultOK(hash, "")
}

func (merk *MerkleApp) Query(query []byte) tmsp.Result {
	if len(query) == 0 {
		return tmsp.ErrEncodingError.SetLog("Query cannot be zero length")
	}
	typeByte := query[0]
	query = query[1:]
	switch typeByte {
	case 0x01: // Size
		size := merk.tree.Size()
		res := wire.BinaryBytes(size)
		return tmsp.NewResultOK(res, "")
	case 0x02: // Query by key
		key, n, err := wire.GetByteSlice(query)
		if err != nil {
			return tmsp.ErrEncodingError.SetLog(
				Fmt("Error getting key: %v", err.Error()))
		}
		query = query[n:]
		if len(query) != 0 {
			return tmsp.ErrEncodingError.SetLog("Got bytes left over")
		}
		_, value, _ := merk.tree.Get(key)
		if len(value) == 0 {
			return tmsp.NewResult(
				ErrValueNotFound, nil, Fmt("Error no value found for query: %v", query))
		}
		return tmsp.NewResultOK(value, "")
	default:
		return tmsp.ErrUnknownRequest.SetLog(
			Fmt("Unexpected Query type byte %X", typeByte))
	}
}