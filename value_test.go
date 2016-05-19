package sysgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	_ "math"
	_ "reflect"
	"testing"
)

func TestValue1GetSetBit(t *testing.T) {
	assert := assert.New(t)
	
	v1, ok1 := NewValue(1).(*Value1)
	assert.Equal(true, ok1, "val not a Value1")

	assert.NotNil(v1.SetBit(1, Hi), "Should have error, but have nil")

	states := []LogicState{Lo, Hi, Undefined, HiZ}
	for _, s := range states {
		v1.SetBit(0, s)
		b, _ := v1.GetBit(0)
		assert.Equal(s, b, fmt.Sprintf("bit should be %s but is %s", s, b))
		assert.Equal(s, v1.v, fmt.Sprintf("bit should be %s but is %s", s, b))
	}

	// Make sure the range functions for Value1 return errors
	var e error
	_, e = v1.GetBits([]uint{0})
	assert.NotNil(e, "Should have error from Value1.GetBits()")

	_, e = v1.GetBitRange(0, 31)
	assert.NotNil(e, "Should have error from Value1.GetBitRange()")

	e = v1.SetBitRange(0, 31, NewValue(32))
	assert.NotNil(e, "Should have error from Value1.SetBitRange()")
}

func TestValue64GetSetBit(t *testing.T) {
	assert := assert.New(t)

	val, ok := NewValue(37).(*Value64)
	assert.Equal(true, ok, "val not a Value64")
	
	// Set lo/hi first
	val.SetBit(0, Hi)
	b, _ := val.GetBit(0)
	assert.Equal(uint64(0x1), val.bits, "Bit 0 not Hi")
	assert.Equal(Hi, b, "Bit 0 not Hi")

	val.SetBit(36, Hi)
	b, _ = val.GetBit(36)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 36 not Hi")
	assert.Equal(Hi, b, "Bit 36 not Hi")

	val.SetBit(35, HiZ)
	b, _ = val.GetBit(35)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(uint64(0x0800000000), val.hiz, "Bit 35 not Hiz")
	assert.Equal(HiZ, b, "Bit 35 not HiZ")

	val.SetBit(13, Undefined)
	b, _ = val.GetBit(13)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(uint64(0x2000), val.undef, "Bit 13 not Undefined")
	assert.Equal(Undefined, b, "Bit 13 not Undefined")

	val.SetBit(13, Lo)
	b, _ = val.GetBit(13)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 13 not Lo")
	assert.Equal(uint64(0x0), val.undef, "Bit 13 not Lo")
	assert.Equal(uint64(0x800000000), val.hiz, "Bit 13 not Lo")
	assert.Equal(Lo, b, "Bit 13 not Lo")
	
	val.SetBit(36, HiZ)
	b, _ = val.GetBit(36)
	assert.Equal(uint64(0x0000000001), val.bits, "Bit 36 not HiZ")
	assert.Equal(uint64(0x1800000000), val.hiz, "Bit 36 not HiZ")
	assert.Equal(HiZ, b, "Bit 36 not HiZ")

	val.SetBit(0, Lo)
	b, _ = val.GetBit(0)
	assert.Equal(uint64(0x0000000000), val.bits, "Bit 0 not Lo")
	assert.Equal(Lo, b, "Bit 0 not Lo")

	t1 := NewValue(8)
	t1.SetBit(0, Hi)
	t1.SetBit(1, Lo)
	t1.SetBit(2, Lo)
	t1.SetBit(3, Hi)
	t1.SetBit(4, HiZ)
	t1.SetBit(5, Undefined)
	t1.SetBit(6, HiZ)
	t1.SetBit(7, Undefined)
	e := val.SetBitRange(16, 23, t1.(*Value64))
	assert.Equal(nil, e, "Error when setting bit range")

	assert.Equal(uint64(0x0000090000), val.bits,  "Incorrect bit range setting")
	assert.Equal(uint64(0x1800500000), val.hiz,   "Incorrect bit range setting")
	assert.Equal(uint64(0x0000a00000), val.undef, "Incorrect bit range setting")

	t2, e := val.GetBitRange(0, 38)
	assert.Equal(true, e != nil, "Should have error in GetBitRange, but none received")
	t2, _ = val.GetBitRange(16, 23)
	t3 := t2.(*Value64)
	assert.Equal(uint64(0x09), t3.bits,  "Incorrect bit range getting")
	assert.Equal(uint64(0x50), t3.hiz,   "Incorrect bit range getting")
	assert.Equal(uint64(0xa0), t3.undef, "Incorrect bit range getting")

	t1.SetBit(0, Undefined)
	t1.SetBit(1, Hi)
	e = val.SetBitRange(16, 23, t1.(*Value64))
	assert.Equal(uint64(0x00000a0000), val.bits,  "Incorrect bit range setting")
	assert.Equal(uint64(0x1800500000), val.hiz,   "Incorrect bit range setting")
	assert.Equal(uint64(0x0000a10000), val.undef, "Incorrect bit range setting")

}

func TestValue1UnaryOps(t *testing.T) {
	assert := assert.New(t)

	v := NewValue(1)

	states := []LogicState{Lo, Hi, Undefined, HiZ}
	for _, s := range states {
		v.SetBit(0, s)
		inv := v.Unary("~")
		var exp LogicState
		switch s {
		case Lo:
			exp = Hi
		case Hi:
			exp = Lo
		default:
			exp = Undefined
		}
		b, _ := inv.GetBit(0)
		assert.Equal(exp, b, "Incorrect inversion for %s", s)
	}
}

func TestValueUnaryOps(t *testing.T) {
	assert := assert.New(t)
	
	var val ValueInterface
	sizes := []uint{48, 61, 212, 247}

	for _, size := range sizes {
		val = NewValue(size)

		switch {
		case size <= 64:
			_, ok := val.(*Value64)
			assert.Equal(true, ok, "Wrong kind")
		case size > 64:
			_, ok := val.(*ValueBig)
			assert.Equal(true, ok, "Wrong kind")
		}

		fmt.Printf("size = %d\n", size)
		val.SetBit(size - 2, Hi)
		inv := val.Unary("~")

		b, _ := inv.GetBit(size - 2)
		assert.Equal(Lo, b, "inv incorrect")
	
		or := val.Unary("|")
		b, _ = or.GetBit(0)
		assert.Equal(uint(1), or.BitLen(), "or has more than 1 bit")
		assert.Equal(Hi, b, "Bit 0 should be Hi")

		nor := val.Unary("~|")
		b, _ = nor.GetBit(0)
		assert.Equal(uint(1), nor.BitLen(), "nor has more than 1 bit")
		assert.Equal(Lo, b, "Bit 0 should be Lo")
	
		and := val.Unary("&")
		b, _ = and.GetBit(0)
		assert.Equal(uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(Lo, b, "Bit 0 should be Lo")

		oddParity := val.Unary("^")
		b, _ = oddParity.GetBit(0)
		assert.Equal(uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(Hi, b, fmt.Sprintf("Bit 0 should be Hi, size = %d", size))
		
		evenParity := val.Unary("~^")
		b, _ = evenParity.GetBit(0)
		assert.Equal(uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(Lo, b, fmt.Sprintf("Bit 0 should be Lo, size = %d", size))

		
		for i := uint(0); i < val.BitLen(); i++ {
			val.SetBit(i, Hi)
		}
		and = val.Unary("&")
		b, _ = and.GetBit(0)
		assert.Equal(uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(Hi, b, "Bit 0 should be Hi")

		nand := val.Unary("~&")
		b, _ = nand.GetBit(0)
		assert.Equal(uint(1), nand.BitLen(), "nand has more than 1 bit")
		assert.Equal(Lo, b, "Bit 0 should be Lo")

		exp := Lo
		notExp := Hi
		if size % 2 != 0 {
			exp = Hi
			notExp = Lo
		}
		oddParity = val.Unary("^")
		b, _ = oddParity.GetBit(0)
		assert.Equal(uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(exp, b, "Bit 0 should be Lo")
		
		evenParity = val.Unary("~^")
		b, _ = evenParity.GetBit(0)
		assert.Equal(uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(notExp, b, "Bit 0 should be Hi")

	}

}
