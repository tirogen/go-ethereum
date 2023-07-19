// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice.
func packBytesSlice(bytes []byte, l int) ([]byte, error) {
	len, err := packNum(reflect.ValueOf(l))
	if err != nil {
		return []byte{}, err
	}
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...), nil
}

// packElement packs the given reflect value according to the abi specification in
// t.
func packElement(t Type, reflectValue reflect.Value) ([]byte, error) {
	switch t.T {
	case IntTy, UintTy:
		return packNum(reflectValue)
	case StringTy:
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len())
	case AddressTy:
		if reflectValue.Type() == reflect.TypeOf("") {
			addr := reflectValue.String()
			reflectValue = reflect.ValueOf(common.HexToAddress(addr))
			if addr != "0x0000000000000000000000000000000000000000" {
				if reflectValue.Interface().(common.Address).Hex() != addr {
					return []byte{}, fmt.Errorf("Could not pack element, invalid address: %v", addr)
				}
			}
		}

		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}

		return common.LeftPadBytes(reflectValue.Bytes(), 32), nil
	case BoolTy:
		if reflectValue.Bool() {
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		return math.PaddedBigBytes(common.Big0, 32), nil
	case BytesTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		if reflectValue.Type() != reflect.TypeOf([]byte{}) {
			return []byte{}, errors.New("Bytes type is neither slice nor array")
		}
		return packBytesSlice(reflectValue.Bytes(), reflectValue.Len())
	case FixedBytesTy, FunctionTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		return common.RightPadBytes(reflectValue.Bytes(), 32), nil
	default:
		return []byte{}, fmt.Errorf("Could not pack element, unknown type: %v", t.T)
	}
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation.
func packNum(value reflect.Value) ([]byte, error) {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return math.U256Bytes(new(big.Int).SetUint64(value.Uint())), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return math.U256Bytes(big.NewInt(value.Int())), nil
	case reflect.Ptr:
		return math.U256Bytes(new(big.Int).Set(value.Interface().(*big.Int))), nil
	case reflect.String:
		bn, ok := new(big.Int).SetString(value.Interface().(string), 10)
		if !ok {
			return []byte{}, fmt.Errorf("Could not pack number, invalid string: %v", value.Interface().(string))
		}
		return math.U256Bytes(bn), nil
	default:
		return []byte{}, fmt.Errorf("Could not pack number, unknown kind: %v", kind)
	}
}
