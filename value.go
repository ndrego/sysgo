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
}

// Value struct capable of representing >64 bits 
type ValueBig struct {
	numBits uint
	bits *big.Int
	hiz *big.Int
	undef *big.Int
}

type ValueInterface interface {
	BitLen() uint
	combine(ValueInterface) error
	Concat(ValueInterface) ValueInterface
	GetBit(uint) LogicState
	GetBits([]uint) ValueInterface
	GetBitRange(low, high uint) ValueInterface
	SetBit(uint, LogicState) error
	SetBitRange(uint, uint, ValueInterface) error
	Text(rune) string
	Unary(string) ValueInterface
	Binary(string, ValueInterface) ValueInterface
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
	Z := NewValue(newNumBits).(*Value64)
	mask := uint64(1 << newNumBits) - 1
	Z.bits = (X.bits >> low) & mask
	Z.hiz = (X.hiz >> low) & mask
	Z.undef = (X.undef >> low) & mask

	return Z
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
	if newNumBits <= 64 {
		t := new(big.Int)
		Z := NewValue(newNumBits).(*Value64)
		mask := uint64(1 << newNumBits) - 1
		Z.bits  = t.Rsh(X.bits,  low).Uint64() & mask
		t.SetUint64(uint64(0))
		Z.hiz   = t.Rsh(X.hiz,   low).Uint64() & mask
		t.SetUint64(uint64(0))
		Z.undef = t.Rsh(X.undef, low).Uint64() & mask
		return Z
	} else {
		Z := NewValue(newNumBits).(*ValueBig)
		one := new(big.Int)
		one.SetUint64(uint64(1))
		mask := new(big.Int)
		mask.Sub(mask.Lsh(one, newNumBits), one)

		Z.bits.And(Z.bits.Rsh(X.bits, low), mask)
		Z.hiz.And(Z.hiz.Rsh(X.hiz, low), mask)
		Z.undef.And(Z.undef.Rsh(X.undef, low), mask)
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
		n = new(Value64)
		n.numBits = 1
		switch v.v {
		case Lo, Hi:
			n.bits = uint64(v.v)
		case HiZ:
			n.hiz = uint64(1)
		case Undefined:
			n.undef = uint64(1)
		}
	case *Value64:
		n = v
	case *ValueBig:
		m := new(big.Int)
		t := new(big.Int)
		m.SetUint64(uint64(1 << 64 - 1))
		n = new(Value64)
		n.numBits = v.BitLen()
		n.bits  = t.And(m, v.bits).Uint64()
		n.hiz   = t.And(m, v.hiz).Uint64()
		n.undef = t.And(m, v.undef).Uint64()
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
		n = new(ValueBig)
		n.bits = new(big.Int)
		n.hiz = new(big.Int)
		n.undef = new(big.Int)
		switch v.v {
		case Lo, Hi:
			n.bits.SetBit(n.bits, 0, uint(v.v))
		case HiZ:
			n.hiz.SetBit(n.hiz, 0, uint(1))
		case Undefined:
			n.undef.SetBit(n.undef, 0, uint(1))
		}
	case *Value64:
		n = new(ValueBig)
		n.bits = new(big.Int)
		n.hiz = new(big.Int)
		n.undef = new(big.Int)
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
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
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
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
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
				Z.SetBit(0, Z.GetBit(0).Unary('~'))
			}
		}
		return Z
	case "&", "~&":
		Z := NewValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.SetBit(0, Undefined)
		} else {
			mask := new(big.Int)
			one := big.NewInt(int64(1))
			mask.Sub(mask.Lsh(one, uint(X.numBits)), one)
			if X.bits.Cmp(mask) == 0 {
				Z.SetBit(0, Hi)
			} else {
				Z.SetBit(0, Lo)
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

func (X *Value1) Binary(op string, Y ValueInterface) (Z ValueInterface) {
	return
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
