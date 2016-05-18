package sysgo

import (
	"fmt"
	"math/big"
	"sort"
)

var ParityTable256 = [...]int{
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0}

type UintSlice []uint

func (A UintSlice) Len() int           { return len(A) }
func (A UintSlice) Swap(i, j int)      { A[i], A[j] = A[j], A[i] }
func (A UintSlice) Less(i, j int) bool { return A[i] < A[j] }
	
type Value64 struct {
	numBits uint
	bits uint64
	hiz uint64
	undef uint64
}

type ValueBig struct {
	numBits uint
	bits *big.Int
	hiz *big.Int
	undef *big.Int
}

type ValueInterface interface {
	BitLen() uint
	GetBit(uint) (LogicState, error)
	GetBits([]uint) (ValueInterface, error)
	GetBitRange(low, high uint) (ValueInterface, error)
	SetBit(uint, LogicState) error
	SetBitRange(uint, uint, ValueInterface) error
	Unary(string) ValueInterface
	Binary(string, ValueInterface) ValueInterface
}


func (X *Value64) copy() (Z *Value64) {
	Z = NewValue(X.BitLen()).(*Value64)
	Z.bits = X.bits
	Z.hiz = X.hiz
	Z.undef = X.undef

	return
}

func (X *ValueBig) copy() (Z *ValueBig) {
	Z = NewValue(X.BitLen()).(*ValueBig)
	Z.bits.SetBytes(X.bits.Bytes())
	Z.hiz.SetBytes(X.hiz.Bytes())
	Z.undef.SetBytes(X.undef.Bytes())

	return
}

func (X *Value64) BitLen() uint {
	return X.numBits
}

func (X *ValueBig) BitLen() uint {
	return X.numBits
}

func (X *Value64) GetBit(b uint) (LogicState, error) {
	if b > (X.numBits - 1) {
		return Undefined, fmt.Errorf("Index (%d) out of bounds.\n", b)
	}

	mask := uint64(1 << b)

	if X.undef & mask != 0 {
		return Undefined, nil
	} else if X.hiz & mask > 0 {
		return HiZ, nil
	} else {
		return LogicState((X.bits >> b) & 0x1), nil
	}
}

func (X *ValueBig) GetBit(b uint) (LogicState, error) {
	if b > (X.numBits - 1) {
		return Undefined, fmt.Errorf("Index (%d) out of bounds.\n", b)
	}

	if X.undef.Bit(int(b)) == 1 {
		return Undefined, nil
	} else if X.hiz.Bit(int(b)) == 1 {
		return HiZ, nil
	} else {
		return LogicState(X.bits.Bit(int(b))), nil
	}
}

func (X *Value64) GetBits(bits []uint) (ValueInterface, error) {
	Z := NewValue(uint(len(bits)))

	sort.Sort(UintSlice(bits))
	for i, b := range bits {
		bitVal, err := X.GetBit(b)
		if err != nil {
			return nil, err
		}
		Z.SetBit(uint(i), bitVal)
	}
	return Z, nil
}

func (X *ValueBig) GetBits(bits []uint) (ValueInterface, error) {
	Z := NewValue(uint(len(bits)))

	sort.Sort(UintSlice(bits))
	for i, b := range bits {
		bitVal, err := X.GetBit(b)
		if err != nil {
			return nil, err
		}
		Z.SetBit(uint(i), bitVal)
	}
	return Z, nil
}

func (X *Value64) GetBitRange(low, high uint) (ValueInterface, error) {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		return nil, fmt.Errorf("low (%d) index out of bounds.\n", low)
	}
	if high > (X.numBits - 1) {
		return nil, fmt.Errorf("high (%d) index out of bounds.\n", high)
	}
	newNumBits := high - low + 1
	Z := NewValue(newNumBits).(*Value64)
	mask := uint64(1 << newNumBits) - 1
	Z.bits = (X.bits >> low) & mask
	Z.hiz = (X.hiz >> low) & mask
	Z.undef = (X.undef >> low) & mask

	return Z, nil
}

func (X *ValueBig) GetBitRange(low, high uint) (ValueInterface, error) {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		return nil, fmt.Errorf("low (%d) index out of bounds.\n", low)
	}
	if high > (X.numBits - 1) {
		return nil, fmt.Errorf("high (%d) index out of bounds.\n", high)
	}
	newNumBits := high - low + 1
	if newNumBits <= 64 {
		t := new(big.Int)
		Z := NewValue(newNumBits).(*Value64)
		mask := uint64(1 << newNumBits) - 1
		Z.bits  = t.Rsh(X.bits,  low).Uint64() & mask
		t.SetUint64(uint64(0))
		Z.hiz   = t.Rsh(X.hiz,   low).Uint64() & mask
		t.SetUint64(uint64(0))
		Z.undef = t.Rsh(X.undef, low).Uint64() & mask
		return Z, nil
	} else {
		Z := NewValue(newNumBits).(*ValueBig)
		one := new(big.Int)
		one.SetUint64(uint64(1))
		mask := new(big.Int)
		mask.Sub(mask.Lsh(one, newNumBits), one)

		Z.bits.And(Z.bits.Rsh(X.bits, low), mask)
		Z.hiz.And(Z.hiz.Rsh(X.hiz, low), mask)
		Z.undef.And(Z.undef.Rsh(X.undef, low), mask)
		return Z, nil
	}
}

func (X *Value64) SetBit(b uint, v LogicState) error {
	if b > (X.numBits - 1) {
		return fmt.Errorf("Index (%d) out of bounds.\n", b)
	}

	mask := uint64(1 << b)
	
	switch v {
	case Undefined:
		X.bits  &= ^mask
		X.hiz   &= ^mask
		X.undef |=  mask
	case HiZ:
		X.bits  &= ^mask
		X.hiz   |=  mask
		X.undef &= ^mask
	case Hi:
		X.bits  |=  mask
		X.hiz   &= ^mask
		X.undef &= ^mask
	case Lo:
		X.bits  &= ^mask
		X.hiz   &= ^mask
		X.undef &= ^mask
	}
	return nil
}

func (X *ValueBig) SetBit(b uint, v LogicState) error {
	if b > (X.numBits - 1) {
		return fmt.Errorf("Index (%d) out of bounds.\n", b)
	}

	mask := new(big.Int)
	mask.SetBit(mask, int(b), 1)
	
	switch v {
	case Undefined:
		X.undef.Or(X.undef, mask)
		X.hiz.AndNot(X.hiz, mask)
	case HiZ:
		X.hiz.Or(X.hiz, mask)
		X.undef.AndNot(X.undef, mask)
	case Hi:
		X.bits.Or(X.bits, mask)
		X.hiz.AndNot(X.hiz, mask)
		X.undef.AndNot(X.undef, mask)
	case Lo:
		X.bits.AndNot(X.bits, mask)
		X.hiz.AndNot(X.hiz, mask)
		X.undef.AndNot(X.undef, mask)
	}
	return nil
}

// Sets a bit range within X. X[low] will get set to v[0] while
// X[high] gets set to v[high - low]. If low > high, they are automatically
// swapped such that high > low always.
func (X *Value64) SetBitRange(low, high uint, v ValueInterface) error {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		return fmt.Errorf("low (%d) index out of bounds.\n", low)
	}
	if high > (X.numBits - 1) {
		return fmt.Errorf("high (%d) index out of bounds.\n", high)
	}
	numBits := high - low + 1
	if numBits != v.BitLen() {
		return fmt.Errorf("Number of bits specified by low (%d), high (%d) = %d does not equal number of bits passed in (%d).", low, high, numBits, v.BitLen())
	}

	// Clear out the specified range of bits then OR in the new bits
	mask := uint64(1 << numBits - 1) << low
	var n *Value64
	switch v := v.(type) {
	case *Value64:
		n = v
	case *ValueBig:
		n = new(Value64)
		n.numBits = v.BitLen()
		n.bits  = v.bits.Uint64()
		n.hiz   = v.hiz.Uint64()
		n.undef = v.undef.Uint64()
	}
	X.bits  = (X.bits  & ^mask) | ((n.bits  << low) & mask)
	X.hiz   = (X.hiz   & ^mask) | ((n.hiz   << low) & mask)
	X.undef = (X.undef & ^mask) | ((n.undef << low) & mask)

	return nil
}

// Sets a bit range within X. X[low] will get set to v[0] while
// X[high] gets set to v[high - low]. If low > high, they are automatically
// swapped such that high > low always.
func (X *ValueBig) SetBitRange(low, high uint, v ValueInterface) error {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		return fmt.Errorf("low (%d) index out of bounds.\n", low)
	}
	if high > (X.numBits - 1) {
		return fmt.Errorf("high (%d) index out of bounds.\n", high)
	}
	numBits := high - low + 1
	if numBits != v.BitLen() {
		return fmt.Errorf("Number of bits specified by low (%d), high (%d) = %d does not equal number of bits passed in (%d).", low, high, numBits, v.BitLen())
	}

	var n *ValueBig
	switch v := v.(type) {
	case *Value64:
		n = new(ValueBig)
		n.bits.SetUint64(v.bits)
		n.hiz.SetUint64(v.hiz)
		n.undef.SetUint64(v.undef)
	case *ValueBig:
		n = v
	}
		
	// Clear out the specified range of bits then OR in the new bits
	mask := new(big.Int)
	one := new(big.Int)
	one.SetUint64(uint64(1))
	mask.Lsh(mask.Sub(mask.Lsh(one, numBits), one), low)

	X.bits.AndNot( X.bits,  mask)
	X.hiz.AndNot(  X.hiz,   mask)
	X.undef.AndNot(X.undef, mask)

	t := new(big.Int)
	X.bits.Or( X.bits,  t.Lsh(n.bits,  low).And(t, mask))
	t.SetInt64(int64(0))
	X.hiz.Or(  X.hiz,   t.Lsh(n.hiz,   low).And(t, mask))
	t.SetInt64(int64(0))
	X.undef.Or(X.undef, t.Lsh(n.undef, low).And(t, mask))

	return nil
}

// Performs a Unary operation on X and returns a new Value. Legal
// Unary operations are: ~ (bit-wise invert) and all reduction operators:
// | (bitwise OR), ~| (bitwise NOR), & (bitwise AND), ~& (bitwise NAND),
// ^ (bitwise XOR / even parity), ~^ (bitwise XNOR / odd parity).
func (X *Value64) Unary(op string) ValueInterface {
	mask := uint64(1 << X.numBits - 1)

	switch op {
	case "~":
		Z := X.copy()
		Z.bits = ^Z.bits & mask
		return Z
	case "|", "~|":
		Z := NewValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.SetBit(0, Undefined)
		} else {
			if X.bits & mask != 0 {
				Z.SetBit(0, Hi)
			} else {
				Z.SetBit(0, Lo)
			}

			if op == "~|" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z
	case "&", "~&":
		Z := NewValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.SetBit(0, Undefined)
		} else {
			if X.bits == mask {
				Z.SetBit(0, Hi)
			} else {
				Z.SetBit(0, Lo)
			}

			if op == "~&" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z
	case "^", "~^":
		Z := NewValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.SetBit(0, Undefined)
		} else {
			// XOR each byte and then look up the resultant
			// value in the parity look up table
			numBytes := X.numBits / 8
			extraBits := X.numBits % 8

			var v, m uint8
			start := int(numBytes) - 1
			if extraBits == 0 {
				m = uint8(0xff)
			} else {
				start += 1
				m = uint8(1 << extraBits - 1)
			}
			for i := start; i >= 0; i-- {
				s := uint8(i * 8)
				if i == start {
					v = uint8(X.bits >> s) & m
				} else {
					v ^= uint8((X.bits >> s) & 0xff)
				}
			}
			Z.SetBit(0, LogicState(uint8(ParityTable256[v])))

			if op == "~^" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z
		
	default:
		return X
	}

}

// Performs a Unary operation on X and returns a new Value. Legal
// Unary operations are: ~ (bit-wise invert) and all reduction operators:
// | (bitwise OR), ~| (bitwise NOR), & (bitwise AND), ~& (bitwise NAND),
// ^ (bitwise XOR / even parity), ~^ (bitwise XNOR / odd parity).
func (X *ValueBig) Unary(op string) ValueInterface {
	var zero big.Int
	switch op {
	case "~":
		Z := X.copy()
		Z.bits.Not(Z.bits)
		return Z
	case "|", "~|":
		Z := NewValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		} else {
			if X.bits.Cmp(&zero) != 0 {
				Z.SetBit(0, Hi)
			} else {
				Z.SetBit(0, Lo)
			}
			if op == "~|" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z
	case "&", "~&":
		Z := NewValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		} else {
			var mask big.Int
			mask.Sub(mask.Exp(big.NewInt(2), big.NewInt(int64(X.numBits)), nil), big.NewInt(1))
			if X.bits.Cmp(&mask) == 0 {
				Z.SetBit(0, Hi)
			} else {
				Z.SetBit(0, Lo)
			}
			if op == "~&" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z
	case "^", "~^":
		Z := NewValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		} else {
			// XOR each byte and then look up the resultant
			// value in the parity look up table
			extraBits := X.numBits % 8
			var v, m uint8
			if extraBits == 0 {
				m = uint8(0xff)
			} else {
				m = uint8(1 << extraBits - 1)
			}

			b := X.bits.Bytes()
			for i, by := range b {
				if i == 0 {
					v = uint8(by & m)
				} else {
					v ^= uint8(by)
				}
			}
			Z.SetBit(0, LogicState(uint8(ParityTable256[v])))

			if op == "~^" {
				b, _ := Z.GetBit(0)
				Z.SetBit(0, b.Unary('~'))
			}
		}
		return Z

	default:
		return X
	}

}

func minNumBits(X, Y ValueInterface) uint {
	if X.BitLen() > Y.BitLen() {
		return Y.BitLen()
	} else {
		return X.BitLen()
	}
}

func (X *Value64) Binary(op string, Y ValueInterface) (Z ValueInterface) {	
	switch op {
	case "&":
		//Z := NewValue(minNumBits(X, Y))
		
	case "+":
	}

	return
}

func (X *ValueBig) Binary(op string, Y ValueInterface) (Z ValueInterface) {
	return
}

func NewValue(numBits uint) ValueInterface {
	switch {
	case numBits <= 64:
		val := new(Value64)
		val.numBits = numBits
		return val
	case numBits > 64:
		val := new(ValueBig)
		val.numBits = numBits
		val.bits  = new(big.Int)
		val.hiz   = new(big.Int)
		val.undef = new(big.Int)
		return val
	default:
		return nil
	}
}
