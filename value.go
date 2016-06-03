package sysgo

import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

// A single-bit value. We optimize for single bit
// values since there are many wires in a design and want them
// to be as fast as possible.
type Value1 struct {
	v LogicState
}

// Value struct capable of representing 1-64 bits
type Value64 struct {
	numBits uint
	bits uint64
	hiz uint64
	undef uint64
	mask uint64
}

// Value struct capable of representing >64 bits 
type ValueBig struct {
	numBits uint
	bits *big.Int
	hiz *big.Int
	undef *big.Int
	mask *big.Int
}

type ValueInterface interface {
	BitLen() uint
	Concat(ValueInterface) ValueInterface
	GetBit(uint) LogicState
	GetBits([]uint) ValueInterface
	GetBitRange(low, high uint) ValueInterface
	SetBit(uint, LogicState) error
	SetBitRange(uint, uint, ValueInterface) error

	String() string
	Text(rune) string
	
	Unary(string) ValueInterface
	Binary(string, ValueInterface) ValueInterface
	Lsh(uint) ValueInterface
	Rsh(uint) ValueInterface
	IsZero() bool
	HasHiz() bool
	HasUndef() bool

	// Private methods
	combine(ValueInterface) error
	cmpEquality(string, ValueInterface) *Value1
	cmpRelational(string, ValueInterface) *Value1
	binLogical(string, ValueInterface) *Value1
	binBitwise(string, ValueInterface) ValueInterface
	// binArithmetic(string, ValueInterface) ValueInterface
}

func (X *Value1) copy() (Z *Value1) {
	Z = NewValue(1).(*Value1)
	Z.v = X.v
	return
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

// Helper conversion functions
func (X *Value1) toValue64(numBits uint) (Y *Value64) {
	Y = new(Value64)
	if numBits == 0 {
		Y.numBits = 1
	} else {
		Y.numBits = numBits
	}
	switch X.v {
	case Lo, Hi:
		Y.bits = uint64(X.v)
	case HiZ:
		Y.hiz = uint64(1)
	case Undefined:
		Y.undef = uint64(1)
	}
	return
}

func (X *Value1) toValueBig(numBits uint) (Y *ValueBig) {
	Y = new(ValueBig)
	if numBits == 0 {
		Y.numBits = 1
	} else {
		Y.numBits = numBits
	}
	Y.bits = new(big.Int)
	Y.hiz = new(big.Int)
	Y.undef = new(big.Int)
	switch X.v {
	case Lo, Hi:
		Y.bits.SetBit(Y.bits, 0, uint(X.v))
	case HiZ:
		Y.hiz.SetBit(Y.hiz, 0, uint(1))
	case Undefined:
		Y.undef.SetBit(Y.undef, 0, uint(1))
	}
	return
}

func (X *Value64) toValue1() (Y *Value1) {
	Y = new(Value1)
	Y.v = X.GetBit(0)
	return
}

func (X *Value64) toValueBig(numBits uint) (Y *ValueBig) {
	Y = new(ValueBig)	
	if numBits == 0 {
		Y.numBits = X.numBits
	} else {
		Y.numBits = numBits
	}
	Y.bits = new(big.Int)
	Y.hiz = new(big.Int)
	Y.undef = new(big.Int)
	Y.bits.SetUint64(X.bits)
	Y.hiz.SetUint64(X.hiz)
	Y.undef.SetUint64(X.undef)
	return
}

// Returns a new Value1 object with only the 0th bit
// of X.
func (X *ValueBig) toValue1() (Y *Value1) {
	Y = new(Value1)
	Y.v = X.GetBit(0)
	return
}

// Returns a new Value64 object with only the lower
// 64 bits (0-63) of X.
func (X *ValueBig) toValue64(numBits uint) (Y *Value64) {
	m := uint(0)
	if numBits == 0 {
		if X.numBits > 64 {
			m = 64
		} else {
			m = X.numBits
		}
	} else {
		m = numBits
	}
			
	return X.GetBitRange(0, m).(*Value64)
}

func (X *Value1) BitLen() uint {
	return uint(1)
}

func (X *Value64) BitLen() uint {
	return X.numBits
}

func (X *ValueBig) BitLen() uint {
	return X.numBits
}

// Returns Z = {X, Y}
func (X *Value1) Concat(Y ValueInterface) ValueInterface {
	yNumBits := Y.BitLen()
	Z := NewValue(1 + yNumBits)
	Z.SetBitRange(0, yNumBits - 1, Y)
	Z.SetBit(yNumBits, X.GetBit(0))
	return Z
}

// Returns Z = {X, Y}
func (X *Value64) Concat(Y ValueInterface) ValueInterface {
	yNumBits := Y.BitLen()
	Z := NewValue(X.numBits + yNumBits)
	Z.SetBitRange(0, yNumBits - 1, Y)
	Z.SetBitRange(yNumBits, yNumBits + X.numBits - 1, X)
	return Z
}

// Returns Z = {X, Y}
func (X *ValueBig) Concat(Y ValueInterface) ValueInterface {
	yNumBits := Y.BitLen()
	Z := NewValue(X.numBits + yNumBits)
	Z.SetBitRange(0, yNumBits - 1, Y)
	Z.SetBitRange(yNumBits, yNumBits + X.numBits - 1, X)
	return Z
}
	
func (X *Value1) GetBit(b uint) LogicState {
	if b != 0 {
		fmt.Printf("Index (%d) out of bounds.\n", b)
		return Undefined
	}
	return X.v
}

func (X *Value64) GetBit(b uint) LogicState {
	if b > (X.numBits - 1) {
		fmt.Printf("Index (%d) out of bounds.\n", b)
		return Undefined
	}

	mask := uint64(1 << b)

	if X.undef & mask != 0 {
		return Undefined
	} else if X.hiz & mask > 0 {
		return HiZ
	} else {
		return LogicState((X.bits >> b) & 0x1)
	}
}

func (X *ValueBig) GetBit(b uint) LogicState {
	if b > (X.numBits - 1) {
		fmt.Printf("Index (%d) out of bounds.\n", b)
		return Undefined
	}

	if X.undef.Bit(int(b)) == 1 {
		return Undefined
	} else if X.hiz.Bit(int(b)) == 1 {
		return HiZ
	} else {
		return LogicState(X.bits.Bit(int(b)))
	}
}

func (X *Value1) GetBits(bits []uint) ValueInterface {
	fmt.Printf("Single bit value does not support GetBits().\n")
	return nil 
}

func (X *Value64) GetBits(bits []uint) ValueInterface {
	Z := NewValue(uint(len(bits)))

	sort.Sort(UintSlice(bits))
	for i, b := range bits {
		Z.SetBit(uint(i), X.GetBit(b))
	}
	return Z
}

func (X *ValueBig) GetBits(bits []uint) ValueInterface {
	Z := NewValue(uint(len(bits)))

	sort.Sort(UintSlice(bits))
	for i, b := range bits {
		Z.SetBit(uint(i), X.GetBit(b))
	}
	return Z
}

func (X *Value1) GetBitRange(low, high uint) ValueInterface {
	fmt.Printf("Single bit value does not support GetBits().\n")
	return nil 
}

func (X *Value64) GetBitRange(low, high uint) ValueInterface {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		fmt.Printf("low (%d) index out of bounds.\n", low)
		return nil 
	}
	if high > (X.numBits - 1) {
		fmt.Printf("high (%d) index out of bounds.\n", high)
		return nil 
	}
	newNumBits := high - low + 1

	if newNumBits == 1 {
		Z := NewValue(1).(*Value1)
		Z.SetBit(0, X.GetBit(low))
		return Z
	} else {
		Z := NewValue(newNumBits).(*Value64)
		Z.bits = (X.bits >> low) & Z.mask
		Z.hiz = (X.hiz >> low) & Z.mask
		Z.undef = (X.undef >> low) & Z.mask
		return Z
	}
}

func (X *ValueBig) GetBitRange(low, high uint) ValueInterface {
	if low > high {
		high, low = low, high
	}
	if low > (X.numBits - 1) {
		fmt.Printf("low (%d) index out of bounds.\n", low)
		return nil 
	}
	if high > (X.numBits - 1) {
		fmt.Printf("high (%d) index out of bounds.\n", high)
		return nil
	}
	newNumBits := high - low + 1
	if newNumBits == 1 {
		Z := NewValue(1).(*Value1)
		Z.SetBit(0, X.GetBit(low))
		return Z
	} else if newNumBits <= 64 {
		t := new(big.Int)
		Z := NewValue(newNumBits).(*Value64)
		Z.bits  = t.Rsh(X.bits,  low).Uint64() & Z.mask
		t.SetUint64(uint64(0))
		Z.hiz   = t.Rsh(X.hiz,   low).Uint64() & Z.mask
		t.SetUint64(uint64(0))
		Z.undef = t.Rsh(X.undef, low).Uint64() & Z.mask
		return Z
	} else {
		Z := NewValue(newNumBits).(*ValueBig)

		Z.bits.And(Z.bits.Rsh(X.bits, low), Z.mask)
		Z.hiz.And(Z.hiz.Rsh(X.hiz, low), Z.mask)
		Z.undef.And(Z.undef.Rsh(X.undef, low), Z.mask)
		return Z
	}
}

func (X *Value1) SetBit(b uint, v LogicState) error {
	if b != 0 {
		return fmt.Errorf("Index (%d) out of bounds.\n", b)
	}
	X.v = v
	return nil
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
		X.bits.AndNot( X.bits,  mask)
		X.hiz.AndNot(  X.hiz,   mask)
		X.undef.Or(    X.undef, mask)
	case HiZ:
		X.bits.AndNot( X.bits,  mask)
		X.hiz.Or(      X.hiz,   mask)
		X.undef.AndNot(X.undef, mask)
	case Hi:
		X.bits.Or(     X.bits,  mask)
		X.hiz.AndNot(  X.hiz,   mask)
		X.undef.AndNot(X.undef, mask)
	case Lo:
		X.bits.AndNot( X.bits,  mask)
		X.hiz.AndNot(  X.hiz,   mask)
		X.undef.AndNot(X.undef, mask)
	}
	return nil
}

func (X *Value1) SetBitRange(low, high uint, v ValueInterface) error {
	return fmt.Errorf("Single bit value does not support SetBitRange().\n")
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
	case *Value1:
		n = v.toValue64(numBits)
	case *Value64:
		n = v
	case *ValueBig:
		n = v.toValue64(numBits)
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
	case *Value1:
		n = v.toValueBig(X.numBits)
	case *Value64:
		n = v.toValueBig(X.numBits)
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

func (X *Value1) Unary(op string) ValueInterface {
	Z := X.copy()
	if strings.HasPrefix(op, "~") {
		switch X.v {
		case Lo:
			Z.v = Hi
		case Hi:
			Z.v = Lo
		default:
			Z.v = Undefined
		}
	}

	return Z
}

// Performs a Unary operation on X and returns a new Value. Legal
// Unary operations are: ~ (bit-wise invert) and all reduction operators:
// | (bitwise OR), ~| (bitwise NOR), & (bitwise AND), ~& (bitwise NAND),
// ^ (bitwise XOR / even parity), ~^ (bitwise XNOR / odd parity).
func (X *Value64) Unary(op string) ValueInterface {
	switch op {
	case "~":
		Z := X.copy()
		Z.bits = X.mask &^ Z.bits &^ (Z.hiz | Z.undef)
		return Z
	case "|", "~|":
		Z := NewValue(1)
		if X.bits & X.mask != 0 {
			Z.SetBit(0, Hi)
		} else if X.hiz & X.mask != 0 || X.undef & X.mask != 0 {
			Z.SetBit(0, Undefined)
		}
		if op == "~|" {
			Z.SetBit(0, Z.GetBit(0).Unary('~'))
		}
		return Z
	case "&", "~&":
		Z := NewValue(1)
		if X.hiz & X.mask != 0 || X.undef & X.mask != 0 {
			Z.SetBit(0, Undefined)
		} else {
			if X.bits == X.mask {
				Z.SetBit(0, Hi)
			}
			if op == "~&" {
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
			}
		}
		return Z
	case "^", "~^":
		Z := NewValue(1)
		if X.hiz & X.mask != 0 || X.undef & X.mask != 0 {
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
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
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
		var m, t1 big.Int
		one := big.NewInt(int64(1))
		m.Sub(m.Lsh(one, uint(X.numBits)), one)
		t1.Or(Z.hiz, Z.undef)
		Z.bits.AndNot(Z.bits.Not(Z.bits), &t1).And(Z.bits, &m)
		return Z
	case "|", "~|":
		Z := NewValue(1)
		if X.bits.Cmp(&zero) != 0 {
			Z.SetBit(0, Hi)
		} else if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		}
		if op == "~|" {
			Z.SetBit(0, Z.GetBit(0).Unary('~'))
		}
		return Z
	case "&", "~&":
		Z := NewValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		} else {
			if X.bits.Cmp(X.mask) == 0 {
				Z.SetBit(0, Hi)
			}
			if op == "~&" {
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
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
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
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

// Legal operators:
// Arithmetic: "+", "-", "*", "/", "**", "%"
// Logical: "&&", "||"
// Relational: ">", "<", ">=", "<="
// Equality: "==", "!=", "===", "!=="
// Bitwise: "&", "|", "^", "(^~ or ~^)"
// Shift: "<<", ">>", "<<<", ">>>"
// http://web.engr.oregonstate.edu/~traylor/ece474/lecture_verilog/beamer/verilog_operators.pdf

func (X *Value1) Lsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	Z := NewValue(n + 1)
	Z.SetBit(n, X.GetBit(0))
	return Z
}

func (X *Value64) Lsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	Z := NewValue(X.numBits + n)
	switch Z := Z.(type) {
	case *Value64:
		Z.bits  = X.bits  << n
		Z.hiz   = X.hiz   << n
		Z.undef = X.undef << n
	case *ValueBig:
		Z.bits.SetUint64( X.bits)
		Z.hiz.SetUint64(  X.hiz)
		Z.undef.SetUint64(X.undef)
		Z.bits.Lsh( Z.bits,  n)
		Z.hiz.Lsh(  Z.hiz,   n)
		Z.undef.Lsh(Z.undef, n)
	}
	return Z
}

func (X *ValueBig) Lsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	Z := X.copy()
	Z.numBits = X.numBits + n
	Z.bits.Lsh( Z.bits,  n)
	Z.hiz.Lsh(  Z.hiz,   n)
	Z.undef.Lsh(Z.undef, n)

	return Z
}

func (X *Value1) Rsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	// Right-shifting a 1-bit value always results in 0
	return NewValue(1)
}

func (X *Value64) Rsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	Z := X.copy()
	Z.bits  >>= n
	Z.hiz   >>= n
	Z.undef >>= n
	return Z
}

func (X *ValueBig) Rsh(n uint) ValueInterface {
	if n == uint(0) {
		return X.copy()
	}
	Z := X.copy()
	Z.bits.Rsh( Z.bits,  n)
	Z.hiz.Rsh(  Z.hiz,   n)
	Z.undef.Rsh(Z.undef, n)

	return Z
}

func (X *Value1) cmpEquality(op string, Y ValueInterface) (Z *Value1) {
	switch Y := Y.(type) {
	case *Value1:
		Z = NewValue(1).(*Value1)
		switch op {
		case "==", "!=":
			if (X.v < HiZ) && (Y.v < HiZ) {
				if X.v == Y.v {
					Z.SetBit(0, Hi)
				}
			} else {
				Z.SetBit(0, Undefined)
			}
		case "===", "!==":
			if X.v == Y.v {
				Z.SetBit(0, Hi)
			}
		}
		if strings.HasPrefix(op, "!") {
			Z.SetBit(0, Z.GetBit(0).Unary('~'))
		}
	case *Value64:
		Z = X.toValue64(Y.BitLen()).cmpEquality(op, Y)
	case *ValueBig:
		Z = X.toValue64(Y.BitLen()).cmpEquality(op, Y)
	}
	return
}

func (X *Value64) cmpEquality(op string, Y ValueInterface) (Z *Value1) {
	switch Y := Y.(type) {
	case *Value1:
		// Equality is symmetric so just flip and return
		Z = Y.cmpEquality(op, X)
	case *Value64:
		Z = NewValue(1).(*Value1)
		switch op {
		case "==", "!=":
			if X.hiz != 0 || Y.hiz != 0 || X.undef != 0 || Y.undef != 0 {
				Z.SetBit(0, Undefined)
			} else {
				if X.bits == Y.bits {
					Z.SetBit(0, Hi)
				}
			}
		case "===", "!==":
			if X.bits == Y.bits && X.hiz == Y.hiz && X.undef == Y.undef {
				Z.SetBit(0, Hi)
			}
		}
		if strings.HasPrefix(op, "!") {
			Z.SetBit(0, Z.GetBit(0).Unary('~'))
		}
	case *ValueBig:
		// Convert this to a ValueBig and then do the comparison
		Z = X.toValueBig(Y.BitLen()).cmpEquality(op, Y)
	}
	return
}

func (X *ValueBig) cmpEquality(op string, Y ValueInterface) (Z *Value1) {
	switch Y := Y.(type) {
	case *Value1, *Value64:
		Z = Y.cmpEquality(op, X)
	case *ValueBig:
		zero := new(big.Int)
		Z = NewValue(1).(*Value1)
		switch op {
		case "==", "!=":
			if X.hiz.Cmp(zero) != 0 || Y.hiz.Cmp(zero) != 0 || X.undef.Cmp(zero) != 0 || Y.undef.Cmp(zero) != 0 {
				Z.SetBit(0, Undefined)
			} else {
				if X.bits.Cmp(Y.bits) == 0 {
					Z.SetBit(0, Hi)
				}
			}
		case "===", "!==":
			if X.bits.Cmp(Y.bits) == 0 && X.hiz.Cmp(Y.hiz) == 0 && X.undef.Cmp(Y.undef) == 0 {
				Z.SetBit(0, Hi)
			}
		}
		if strings.HasPrefix(op, "!") {
			Z.SetBit(0, Z.GetBit(0).Unary('~'))
		}
	}
	return
}

func (X *Value1) cmpRelational(op string, Y ValueInterface) (Z *Value1) {
	switch Y := Y.(type) {
	case *Value1:
		Z = NewValue(1).(*Value1)
		if (X.v < HiZ) && (Y.v < HiZ) {
			switch op {
			case "<":
				if X.v < Y.v {
					Z.SetBit(0, Hi)
				}
			case "<=":
				if X.v <= Y.v {
					Z.SetBit(0, Hi)
				}
			case ">":
				if X.v > Y.v {
					Z.SetBit(0, Hi)
				}
			case ">=":
				if X.v >= Y.v {
					Z.SetBit(0, Hi)
				}
			}
		} else {
			Z.SetBit(0, Undefined)
		}
	case *Value64:
		Z = X.toValue64(Y.BitLen()).cmpRelational(op, Y)
	case *ValueBig:
		Z = X.toValueBig(Y.BitLen()).cmpRelational(op, Y)
	}
	return
}

func (X *Value64) cmpRelational(op string, Y ValueInterface) (Z *Value1) {
	var n *Value64
	switch Y := Y.(type) {
	case *Value1:
		n = Y.toValue64(X.numBits)
	case *Value64:
		n = Y
	case *ValueBig:
		Z = X.toValueBig(Y.BitLen()).cmpRelational(op, Y)
		return
	}
	Z = NewValue(1).(*Value1)
	
	if X.hiz != 0 || n.hiz != 0 || X.undef != 0 || n.undef != 0 {
		Z.SetBit(0, Undefined)
	} else {
		switch op {
		case "<":
			if X.bits < n.bits {
				Z.SetBit(0, Hi)
			}
		case "<=":
			if X.bits <= n.bits {
				Z.SetBit(0, Hi)
			}
		case ">":
			if X.bits > n.bits {
				Z.SetBit(0, Hi)
			}
		case ">=":
			if X.bits >= n.bits {
				Z.SetBit(0, Hi)
			}
		}
	}
	return
}

func (X *ValueBig) cmpRelational(op string, Y ValueInterface) (Z *Value1) {
	var n *ValueBig
	switch Y := Y.(type) {
	case *Value1:
		n = Y.toValueBig(X.numBits)
	case *Value64:
		n = Y.toValueBig(X.numBits)
	case *ValueBig:
		n = Y
	}
	Z = NewValue(1).(*Value1)

	zero := new(big.Int)
	if X.hiz.Cmp(zero) != 0 || n.hiz.Cmp(zero) != 0 || X.undef.Cmp(zero) != 0 || n.undef.Cmp(zero) != 0 {
		Z.SetBit(0, Undefined)
	} else {
		res := X.bits.Cmp(n.bits)
		switch op {
		case "<":
			if res == -1 {
				Z.SetBit(0, Hi)
			}
		case "<=":
			if res <= 0 {
				Z.SetBit(0, Hi)
			}
		case ">":
			if res == 1 {
				Z.SetBit(0, Hi)
			}
		case ">=":
			if res >= 0 {
				Z.SetBit(0, Hi)
			}
		}
	}
	return
}

// Returns true iff X == 0. If X = (Hi | HiZ | Undefined), returns false.
func (X *Value1) IsZero() bool {
	return X.v == Lo
}

// Returns true iff X == 0. If X = has any non-zero bits (including HiZ or
// Undefined) the return value is false.
func (X *Value64) IsZero() bool {
	return X.bits == uint64(0) && X.hiz == uint64(0) && X.undef == uint64(0)
}

// Returns true iff X == 0. If X = has any non-zero bits (including HiZ or
// Undefined) the return value is false.
func (X *ValueBig) IsZero() bool {
	var z big.Int
	return X.bits.Cmp(&z) == 0 && X.hiz.Cmp(&z) == 0 && X.undef.Cmp(&z) == 0
}

func (X *Value1) HasHiz() bool {
	return X.v == HiZ
}

func (X *Value64) HasHiz() bool {
	return X.hiz != 0
}

func (X *ValueBig) HasHiz() bool {
	var z big.Int
	return X.hiz.Cmp(&z) != 0
}

func (X *Value1) HasUndef() bool {
	return X.v == Undefined
}

func (X *Value64) HasUndef() bool {
	return X.undef != 0
}

func (X *ValueBig) HasUndef() bool {
	var z big.Int
	return X.undef.Cmp(&z) != 0
}

// Computes logical-AND and logical-OR. So:
// Z = X && Y returns 1'b1 if both X and Y are non-zero and 1'b0 otherwise
// Z = X || Y returns 1'b1 if either X or Y are non-zero
func (X *Value1) binLogical(op string, Y ValueInterface) (Z *Value1) {
	Z = NewValue(1).(*Value1)
	Yv := Y.GetBit(0)

	switch op {
	case "&&":
		if X.v > Hi || Yv > Hi {
			Z.SetBit(0, Undefined)
		} else {		
			if X.v == Hi && Yv == Hi {
				Z.SetBit(0, Hi)
			}
		}
	case "||":
		if X.v == Hi || Yv == Hi {
			Z.SetBit(0, Hi)
		} else if X.v > Hi || Yv > Hi {
			Z.SetBit(0, Undefined)
		}
	}
	return
}

// Computes logical-AND and logical-OR. So:
// Z = X && Y returns 1'b1 if both X and Y are non-zero and 1'b0 otherwise
// Z = X || Y returns 1'b1 if either X or Y are non-zero
func (X *Value64) binLogical(op string, Y ValueInterface) (Z *Value1) {
	Xor := X.Unary("|").(*Value1)
	Yor := Y.Unary("|").(*Value1)
	
	Z = Xor.binLogical(op, Yor)
	return
}

// Computes logical-AND and logical-OR. So:
// Z = X && Y returns 1'b1 if both X and Y are non-zero and 1'b0 otherwise
// Z = X || Y returns 1'b1 if either X or Y are non-zero
func (X *ValueBig) binLogical(op string, Y ValueInterface) (Z *Value1) {
	Xor := X.Unary("|").(*Value1)
	Yor := Y.Unary("|").(*Value1)

	Z = Xor.binLogical(op, Yor)
	return
}

func (X *Value1) binBitwise(op string, Y ValueInterface) ValueInterface {
	switch Y := Y.(type) {
	case *Value1:
		Z := NewValue(1)
		switch op {
		case "&":
			if X.v > Hi || Y.v > Hi {
				Z.SetBit(0, Undefined)
			} else {
				if X.v == Hi && Y.v == Hi {
					Z.SetBit(0, Hi)
				}
			}
		case "|":
			if X.v == Hi || Y.v == Hi {
				Z.SetBit(0, Hi)
			} else if X.v > Hi || Y.v > Hi {
				Z.SetBit(0, Undefined)
			}
		case "^", "^~", "~^":
			if X.v > Hi || Y.v > Hi {
				Z.SetBit(0, Undefined)
			} else {
				if (X.v == Hi && Y.v == Lo) || (X.v == Lo && Y.v == Hi) {
					Z.SetBit(0, Hi)
				}
				if strings.ContainsRune(op, '~') {
					Z.SetBit(0, Z.GetBit(0).Unary('~'))
				}
			}
		}
		return Z		
	case *Value64:
		return X.toValue64(Y.BitLen()).binBitwise(op, Y)
	case *ValueBig:
		return X.toValueBig(Y.BitLen()).binBitwise(op, Y)
	default:
		return nil
	}
}

func (X *Value64) binBitwise(op string, Y ValueInterface) ValueInterface {
	var n *Value64
	switch Y := Y.(type) {
	case *Value1:
		n = Y.toValue64(X.BitLen())
	case *Value64:
		n = Y
	case *ValueBig:
		return X.toValueBig(Y.BitLen()).binBitwise(op, Y)
	}

	numBits := X.numBits
	if n.numBits > X.numBits {
		numBits = n.numBits
	}
	Z := NewValue(numBits).(*Value64)

	switch op {
	case "&":
		t1 := X.hiz | X.undef
		t2 := n.hiz | n.undef
		Z.undef = (t1 & n.bits) | (t2 & X.bits) | (t1 & t2)
		Z.bits  = (X.bits & n.bits) &^ Z.undef
	case "|":
		t1 := X.hiz | X.undef
		t2 := n.hiz | n.undef
		Z.bits  = X.bits | n.bits
		Z.undef = (t1 &^ n.bits) | (t2 &^ X.bits) | (t1 & t2)
	case "^", "^~", "~^":
		Z.undef = X.undef | n.undef | X.hiz | n.hiz
		Z.bits  = (X.bits ^ n.bits) &^ Z.undef
		if strings.ContainsRune(op, '~') {
			Z = Z.Unary("~").(*Value64)
		}
	}
	return Z
}

func (X *ValueBig) binBitwise(op string, Y ValueInterface) ValueInterface {
	var n *ValueBig
	switch Y := Y.(type) {
	case *Value1:
		n = Y.toValueBig(X.BitLen())
	case *Value64:
		n = Y.toValueBig(X.BitLen())
	case *ValueBig:
		n = Y
	}

	numBits := X.numBits
	if n.numBits > X.numBits {
		numBits = n.numBits
	}
	Z := NewValue(numBits).(*ValueBig)

	switch op {
	case "&":
		var t1, t2, t3, t4, t5 big.Int
		t1.Or(X.hiz, X.undef)
		t2.Or(n.hiz, n.undef)
		t3.And(&t1, n.bits)
		t4.And(&t2, X.bits)
		t5.And(&t1, &t2)
		Z.undef.Or(&t3, &t4).Or(Z.undef, &t5)
		Z.bits.And(X.bits, n.bits).AndNot(Z.bits, Z.undef)
	case "|":
		var t1, t2, t3, t4, t5 big.Int
		t1.Or(X.hiz, X.undef)
		t2.Or(n.hiz, n.undef)
		Z.bits.Or(X.bits, n.bits)
		t3.AndNot(&t1, n.bits)
		t4.AndNot(&t2, X.bits)
		t5.And(&t1, &t2)
		Z.undef.Or(&t3, &t4).Or(Z.undef, &t5)
	case "^", "^~", "~^":
		Z.undef.Or(X.undef, n.undef).Or(Z.undef, X.hiz).Or(Z.undef, n.hiz)
		Z.bits.Xor(X.bits, n.bits).AndNot(Z.bits, Z.undef)
		if strings.ContainsRune(op, '~') {
			Z = Z.Unary("~").(*ValueBig)
		}
	}
	return Z
	
}

func (X *Value1) Binary(op string, Y ValueInterface) (Z ValueInterface) {
	switch op {
	case "==", "!=", "===", "!==":
		// Equality operators
		Z = X.cmpEquality(op, Y)
	case "<", "<=", ">", ">=":
		// Relational operators
		Z = X.cmpRelational(op, Y)
	case "&&", "||":
		// Logical operators
		Z = X.binLogical(op, Y)
	case "&", "|", "^", "^~", "~^":
		// Bitwise operators
		Z = X.binBitwise(op, Y)
	default:
		fmt.Printf("WARNING: %s is not a defined operator, returning Undefined.\n", op)
		Z, _ = NewValueString("1'bx")		
	}
	return
}

func (X *Value64) Binary(op string, Y ValueInterface) (Z ValueInterface) {	
	switch op {
	case "==", "!=", "===", "!==":
		// Equality operators
		Z = X.cmpEquality(op, Y)
	case "<", "<=", ">", ">=":
		// Relational operators
		Z = X.cmpRelational(op, Y)
	case "&&", "||":
		// Logical operators
		Z = X.binLogical(op, Y)
	case "&", "|", "^", "^~", "~^":
		// Bitwise operators
		Z = X.binBitwise(op, Y)
	default:
		fmt.Printf("WARNING: %s is not a defined operator, returning Undefined.\n", op)
		Z, _ = NewValueString("1'bx")		
	}
	return
}

func (X *ValueBig) Binary(op string, Y ValueInterface) (Z ValueInterface) {
	switch op {
	case "==", "!=", "===", "!==":
		// Equality operators
		Z = X.cmpEquality(op, Y)
	case "<", "<=", ">", ">=":
		// Relational operators
		Z = X.cmpRelational(op, Y)
	case "&&", "||":
		// Logical operators
		Z = X.binLogical(op, Y)
	case "&", "|", "^", "^~", "~^":
		// Bitwise operators
		Z = X.binBitwise(op, Y)
	default:
		fmt.Printf("WARNING: %s is not a defined operator, returning Undefined.\n", op)
		Z, _ = NewValueString("1'bx")		
	}
	return
}

func (X *Value1) combine(Y ValueInterface) error {
	if Y.BitLen() < 1 {
		return fmt.Errorf("Y has too few bits\n")
	}

	var Yv LogicState
	switch Y := Y.(type) {
	case *Value1:
		Yv = Y.v
	default:
		Yv = Y.GetBit(0)
	}
	if X.v == Undefined || Yv == Undefined || X.v != Yv {
		X.v = Undefined
	} else if X.v == HiZ && (Yv == Lo || Yv == Hi) {
		X.v = Yv
	}
	return nil
}

func (X *Value64) combine(Y ValueInterface) error {
	m := minNumBits(X, Y)

	var Z *Value64
	switch Y := Y.(type) {
	case *Value1:
		z := NewValue(1)
		z.SetBit(0, X.GetBit(0))
		z.combine(Y)
		X.SetBit(0, z.GetBit(0))
		return nil
	case *Value64:
		Z = Y
	case *ValueBig:
		Z = Y.GetBitRange(0, m - 1).(*Value64)
	}
	
	mask := uint64(1 << m) - 1
	// HiZ is just an AND of each HiZ, masked
	// by the overlapping bits.
	c := X.hiz & ^mask
	X.hiz = c | (X.hiz & Z.hiz & mask)
		
	// Undef is an OR of the Undefined's plus
	// any bits that don't match up between the two
	X.undef |= (Z.undef & mask)
	X.undef |= ((X.bits ^ Z.bits) & mask)
		
	// To compute the actual bit value, we need to
	// clear any bits that have a HiZ or Undef
	either := X.hiz | X.undef
	X.bits &= ^either

	return nil
}		

func (X *ValueBig) combine(Y ValueInterface) error {
	//m := minNumBits(X, Y)

	var n *ValueBig
	switch Y := Y.(type) {
	case *Value1:
		z := NewValue(1)
		z.SetBit(0, X.GetBit(0))
		z.combine(Y)
		X.SetBit(0, z.GetBit(0))
		return nil
	case *Value64:
		n = new(ValueBig)
		n.numBits = Y.numBits
		n.bits.SetUint64(Y.bits)
		n.hiz.SetUint64(Y.hiz)
		n.undef.SetUint64(Y.undef)
	case *ValueBig:
		n = Y
	}

	return nil
}

func (X *Value1) String() string {
	s := bytes.NewBufferString("1'b")
	s.WriteRune(X.v.Rune())
	return s.String()
}

func (X *Value64) String() string {
	s := bytes.NewBufferString(fmt.Sprintf("%d'b", X.numBits))
	for i := int(X.numBits) - 1; i >= 0; i-- {
		s.WriteRune(X.GetBit(uint(i)).Rune())
	}
	return s.String()
}

func (X *ValueBig) String() string {
	s := bytes.NewBufferString(fmt.Sprintf("%d'b", X.numBits))
	for i := int(X.numBits) - 1; i >= 0; i-- {
		s.WriteRune(X.GetBit(uint(i)).Rune())
	}
	return s.String()
}

// base can be either 'b', 'o' or 'h'
func (X *Value1) Text(base rune) string {
	s := bytes.NewBufferString("1'")
	s.WriteRune(base)
	s.WriteRune(X.v.Rune())
	return s.String()
}

// base can be either 'b', 'o' or 'h'
func (X *Value64) Text(base rune) string {
	s := bytes.NewBufferString(fmt.Sprintf("%d'%s", X.numBits, string(base)))
	n := uint(0)
	fmtSpec := ""
	switch base {
	case 'b':
		n = 1
		fmtSpec = "%b"
	case 'o':
		n = 3
		fmtSpec = "%o"
	case 'h':
		n = 4
		fmtSpec = "%x"
	}

	mask := uint64(1 << n - 1)

	r := make([]string, 0, 1)
	for i := 0; i < int(X.numBits);  {
		sh := uint(i)
		str := fmt.Sprintf(fmtSpec, (X.bits >> sh) & mask)
		if (X.hiz >> sh) & mask != 0 {
			str = "z"
		}
		if (X.undef >> sh) & mask != 0 {
			str = "x"
		}
		r = append(r, str)
		i += int(n)
	}
	for i := len(r) - 1; i >= 0; i-- {
		s.WriteString(r[i])
	}

	return s.String()
}

// base can be either 'b', 'o' or 'h'
func (X *ValueBig) Text(base rune) string {
	s := bytes.NewBufferString(fmt.Sprintf("%d'%s", X.numBits, string(base)))
	n := uint(0)
	fmtSpec := ""
	switch base {
	case 'b':
		n = 1
		fmtSpec = "%b"
	case 'o':
		n = 3
		fmtSpec = "%o"
	case 'h':
		n = 4
		fmtSpec = "%x"
	}

	mask := new(big.Int)
	one := new(big.Int)
	zero := new(big.Int)
	one.SetUint64(uint64(1))
	mask.Sub(mask.Lsh(one, n), one)
	
	r := make([]string, 0, X.numBits / uint(n) + 1)
	for i := 0; i < int(X.numBits);  {
		sh := uint(i)
		t := new(big.Int)
		t.And(t.Rsh(X.bits, sh), mask)
		str := fmt.Sprintf(fmtSpec, t.Uint64())
		t.SetInt64(int64(0))
		t.And(t.Rsh(X.hiz, sh), mask)
		if t.Cmp(zero) != 0 {
			str = "z"
		}
		t.SetInt64(int64(0))
		t.And(t.Rsh(X.undef, sh), mask)
		if t.Cmp(zero) != 0 {
			str = "x"
		}
		r = append(r, str)
		i += int(n)
	}
	for i := len(r) - 1; i >= 0; i-- {
		s.WriteString(r[i])
	}

	return s.String()
}


func NewValue(numBits uint) ValueInterface {
	switch {
	case numBits == 1:
		val := new(Value1)
		return val
	case numBits <= 64:
		val := new(Value64)
		val.numBits = numBits
		val.mask = uint64((1 << numBits) - 1)
		return val
	case numBits > 64:
		val := new(ValueBig)
		val.numBits = numBits
		val.bits  = new(big.Int)
		val.hiz   = new(big.Int)
		val.undef = new(big.Int)
		val.mask  = new(big.Int)

		one := new(big.Int)
		one.SetUint64(uint64(1))
		val.mask.Sub(val.mask.Lsh(one, numBits), one)
		
		return val
	default:
		return nil
	}
}

// Initializes a new Value* based on the contents of s,
// which must be of the form <size>'<signed><radix>value.
// size, signed and radix are all optional. If radix is
// not specified, decimal is assumed.
func NewValueString(s string) (ValueInterface, error) {
	size := 0
	radix := "d"
	// signed := false
	value := []rune("0")

	re, _ := regexp.Compile("(?i)(\\d+)?'(s)?([bhod])?(.*)")
	res := re.FindStringSubmatch(s)
	if res == nil {
		return nil, fmt.Errorf("%s is not a valid value string.", s)
	}

	if len(res[1]) != 0 {
		size, _ = strconv.Atoi(res[1])
	}
	if len(res[2]) != 0 {
		//signed = true
	}
	if len(res[3]) != 0 {
		radix = strings.ToLower(res[3])
	}
		
	value = []rune(strings.ToLower(strings.Replace(res[4], "_", "", -1)))

	hiz := make([]uint, 0, 1)
	undef := make([]uint, 0, 1)

	bitsPerRune := 0
	base := 10
	switch radix {
	case "b":
		bitsPerRune = 1
		base = 2
	case "h":
		bitsPerRune = 4
		base = 16
	case "o":
		bitsPerRune = 3
		base = 8
	}
	for i := len(value) - 1; i >= 0; i-- {
		bitIndex := (len(value) - i - 1) * bitsPerRune
		switch value[i] {
		case 'z':
			for j := 0; j < bitsPerRune; j++ {
				k := uint(bitIndex+j)
				if size == 0 || k < uint(size) {
					hiz = append(hiz, k)
				}
			}
			value[i] = '0'
		case 'x':
			for j := 0; j < bitsPerRune; j++ {
				k := uint(bitIndex+j)
				if size == 0 || k < uint(size) {
					undef = append(undef, k)
				}
			}
			value[i] = '0'
		}
	}

	p := new(big.Int)
	if _, ok := p.SetString(string(value), base); !ok {
		return nil, fmt.Errorf("Couldn't convert %s", s)
	}
	numBits := p.BitLen()

	if size == 0 {
		size = numBits
	} else {
		if numBits > size {
			return nil, fmt.Errorf("Bit length (%d) does not correspond with specified size (%d)", numBits, size)
		}
	}

	newVal := NewValue(uint(size))
	switch newVal := newVal.(type) {
	case *Value1:
		newVal.v = LogicState(uint8(p.Bit(0)))
	case *Value64:
		newVal.bits = p.Uint64()
	case *ValueBig:
		newVal.bits = p
	}

	for _, h := range hiz {
		newVal.SetBit(h, HiZ)
	}
	for _, u := range undef {
		newVal.SetBit(u, Undefined)
	}

	return newVal, nil
}
