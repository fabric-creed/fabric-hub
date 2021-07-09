package util

import (
	"crypto/sha256"
	"encoding/asn1"
	"github.com/fabric-creed/cryptogm/sm3"
	"math/big"

	cb "github.com/fabric-creed/fabric-protos-go/common"
)

type asn1Header struct {
	Number       *big.Int
	PreviousHash []byte
	DataHash     []byte
}

func BlockHeaderBytes(b *cb.BlockHeader) []byte {
	asn1Header := asn1Header{
		PreviousHash: b.PreviousHash,
		DataHash:     b.DataHash,
		Number:       new(big.Int).SetUint64(b.Number),
	}
	result, err := asn1.Marshal(asn1Header)
	if err != nil {
		// Errors should only arise for types which cannot be encoded, since the
		// BlockHeader type is known a-priori to contain only encodable types, an
		// error here is fatal and should not be propogated
		panic(err)
	}
	return result
}

func BlockHeaderHash(b *cb.BlockHeader, isGM bool) []byte {
	if isGM {
		return sm3.SumSM3(BlockHeaderBytes(b))
	}
	sum := sha256.Sum256(BlockHeaderBytes(b))
	return sum[:]
}
