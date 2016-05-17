package sysgo

import (
	_ "fmt"
	"log"
	"math"
	"math/big"
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

type mbv64 struct {
	numBits uint
	bits uint64
	hiz uint64
	undef uint64
}

type mbvBig struct {
	numBits uint
	bits big.Int
	hiz big.Int
	undef big.Int
}

type MultiBitValue interface {
	bitLen() uint
	getBit(uint) LogicValue
	setBit(uint, LogicValue)
	unary(string) MultiBitValue
	binary(string, MultiBitValue) MultiBitValue
}


func (X *mbv64) copy() (Z *mbv64) {
	Z = NewMultiBitValue(X.bitLen()).(*mbv64)
	Z.bits = X.bits
	Z.hiz = X.hiz
	Z.undef = X.undef

	return
}

func (X *mbvBig) copy() (Z *mbvBig) {
	Z = NewMultiBitValue(X.bitLen()).(*mbvBig)
	Z.bits.SetBytes(X.bits.Bytes())
	Z.hiz.SetBytes(X.hiz.Bytes())
	Z.undef.SetBytes(X.undef.Bytes())

	return
}

func (X *mbv64) bitLen() uint {
	return X.numBits
}

func (X *mbvBig) bitLen() uint {
	return X.numBits
}

func (X *mbv64) getBit(b uint) LogicValue {
	if b > (X.numBits - 1) {
		log.Fatal("Index (%d) out of bounds.\n", b)
	}

	mask := uint64(1 << b)

	if X.undef & mask != 0 {
		return Undefined
	} else if X.hiz & mask > 0 {
		return HiZ
	} else {
		return LogicValue((X.bits >> b) & 0x1)
	}
}

func (X *mbvBig) getBit(b uint) LogicValue {
	if b > (X.numBits - 1) {
		log.Fatal("Index (%d) out of bounds.\n", b)
	}

	if X.undef.Bit(int(b)) == 1 {
		return Undefined
	} else if X.hiz.Bit(int(b)) == 1 {
		return HiZ
	} else {
		return LogicValue(X.bits.Bit(int(b)))
	}
}


func (X *mbv64) setBit(b uint, v LogicValue) {
	if b > (X.numBits - 1) {
		log.Fatal("Index (%d) out of bounds.\n", b)
	}

	mask := uint64(1 << b)
	
	switch v {
	case Undefined:
		X.undef |= mask
		X.hiz &= ^mask
	case HiZ:
		X.hiz |= mask
		X.undef &= ^mask
	case Hi:
		X.bits |= mask
		X.hiz &= ^mask
		X.undef &= ^mask
	case Lo:
		X.bits &= ^mask
		X.hiz &= ^mask
		X.undef &= ^mask
	}
}

func (X *mbvBig) setBit(b uint, v LogicValue) {
	if b > (X.numBits - 1) {
		log.Fatal("Index (%d) out of bounds.\n", b)
	}

	var mask big.Int
	mask.SetBit(&mask, int(b), 1)
	
	switch v {
	case Undefined:
		X.undef.Or(&X.undef, &mask)
		X.hiz.AndNot(&X.hiz, &mask)
	case HiZ:
		X.hiz.Or(&X.hiz, &mask)
		X.undef.AndNot(&X.undef, &mask)
	case Hi:
		X.bits.Or(&X.bits, &mask)
		X.hiz.AndNot(&X.hiz, &mask)
		X.undef.AndNot(&X.undef, &mask)
	case Lo:
		X.bits.AndNot(&X.bits, &mask)
		X.hiz.AndNot(&X.hiz, &mask)
		X.undef.AndNot(&X.undef, &mask)
	}
}


func (X *mbv64) unary(op string) MultiBitValue {
	mask := uint64(1 << X.numBits - 1)

	switch op {
	case "~":
		Z := X.copy()
		Z.bits = ^Z.bits & mask
		return Z
	case "|", "~|":
		Z := NewMultiBitValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.setBit(0, Undefined)
		} else {
			if X.bits & mask != 0 {
				Z.setBit(0, Hi)
			} else {
				Z.setBit(0, Lo)
			}

			if op == "~|" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z
	case "&", "~&":
		Z := NewMultiBitValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.setBit(0, Undefined)
		} else {
			if X.bits == mask {
				Z.setBit(0, Hi)
			} else {
				Z.setBit(0, Lo)
			}

			if op == "~&" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z
	case "^", "~^":
		Z := NewMultiBitValue(1)
		if X.hiz & mask != 0 || X.undef & mask != 0 {
			Z.setBit(0, Undefined)
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
			Z.setBit(0, LogicValue(uint8(ParityTable256[v])))

			if op == "~^" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z
		
	default:
		return X
	}

}

func (X *mbvBig) unary(op string) MultiBitValue {
	var zero big.Int
	switch op {
	case "~":
		Z := X.copy()
		Z.bits.Not(&Z.bits)
		return Z
	case "|", "~|":
		Z := NewMultiBitValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.setBit(0, Undefined)
		} else {
			if X.bits.Cmp(&zero) != 0 {
				Z.setBit(0, Hi)
			} else {
				Z.setBit(0, Lo)
			}
			if op == "~|" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z
	case "&", "~&":
		Z := NewMultiBitValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.setBit(0, Undefined)
		} else {
			var mask big.Int
			mask.Sub(mask.Exp(big.NewInt(2), big.NewInt(int64(X.numBits)), nil), big.NewInt(1))
			if X.bits.Cmp(&mask) == 0 {
				Z.setBit(0, Hi)
			} else {
				Z.setBit(0, Lo)
			}
			if op == "~&" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z
	case "^", "~^":
		Z := NewMultiBitValue(1)
		if X.hiz.Cmp(&zero) != 0 || X.undef.Cmp(&zero) != 0 {
			Z.setBit(0, Undefined)
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
			Z.setBit(0, LogicValue(uint8(ParityTable256[v])))

			if op == "~^" {
				Z.setBit(0, Z.getBit(0).Unary('~'))
			}
		}
		return Z

	default:
		return X
	}

}


func (X *mbv64) binary(op string, Y MultiBitValue) (Z MultiBitValue) {	
	switch op {
	case "&":
		Z = NewMultiBitValue(uint(math.Max(float64(X.numBits), float64(Y.bitLen()))))
	case "+":
	}

	return
}

func (X *mbvBig) binary(op string, Y MultiBitValue) (Z MultiBitValue) {
	return
}

func NewMultiBitValue(numBits uint) MultiBitValue {
	switch {
	case numBits <= 64:
		mbv := new(mbv64)
		mbv.numBits = numBits
		return mbv
	case numBits > 64:
		mbv := new(mbvBig)
		mbv.numBits = numBits
		return mbv
	default:
		return nil
	}
}
