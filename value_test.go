package sysgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	_ "math"
	_ "reflect"
	"testing"
)

func TestValueGetSetBit(t *testing.T) {
	val, ok := NewValue(37).(*Value64)
	assert.Equal(t, true, ok, "val not a Value64")
	
	// Set lo/hi first
	val.SetBit(0, Hi)
	b, _ := val.GetBit(0)
	assert.Equal(t, uint64(0x1), val.bits, "Bit 0 not Hi")
	assert.Equal(t, Hi, b, "Bit 0 not Hi")

	val.SetBit(36, Hi)
	b, _ = val.GetBit(36)
	assert.Equal(t, uint64(0x1000000001), val.bits, "Bit 36 not Hi")
	assert.Equal(t, Hi, b, "Bit 36 not Hi")

	val.SetBit(35, HiZ)
	b, _ = val.GetBit(35)
	assert.Equal(t, uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x0800000000), val.hiz, "Bit 35 not Hiz")
	assert.Equal(t, HiZ, b, "Bit 35 not HiZ")

	val.SetBit(13, Undefined)
	b, _ = val.GetBit(13)
	assert.Equal(t, uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x2000), val.undef, "Bit 13 not Undefined")
	assert.Equal(t, Undefined, b, "Bit 13 not Undefined")

	val.SetBit(13, Lo)
	b, _ = val.GetBit(13)
	assert.Equal(t, uint64(0x1000000001), val.bits, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x0), val.undef, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x800000000), val.hiz, "Bit 13 not Lo")
	assert.Equal(t, Lo, b, "Bit 13 not Lo")
	
	val.SetBit(36, HiZ)
	b, _ = val.GetBit(36)
	assert.Equal(t, uint64(0x0000000001), val.bits, "Bit 36 not HiZ")
	assert.Equal(t, uint64(0x1800000000), val.hiz, "Bit 36 not HiZ")
	assert.Equal(t, HiZ, b, "Bit 36 not HiZ")

	val.SetBit(0, Lo)
	b, _ = val.GetBit(0)
	assert.Equal(t, uint64(0x0000000000), val.bits, "Bit 0 not Lo")
	assert.Equal(t, Lo, b, "Bit 0 not Lo")

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
	assert.Equal(t, nil, e, "Error when setting bit range")

	assert.Equal(t, uint64(0x0000090000), val.bits,  "Incorrect bit range setting")
	assert.Equal(t, uint64(0x1800500000), val.hiz,   "Incorrect bit range setting")
	assert.Equal(t, uint64(0x0000a00000), val.undef, "Incorrect bit range setting")

	t2, e := val.GetBitRange(0, 38)
	assert.Equal(t, true, e != nil, "Should have error in GetBitRange, but none received")
	t2, _ = val.GetBitRange(16, 23)
	t3 := t2.(*Value64)
	assert.Equal(t, uint64(0x09), t3.bits,  "Incorrect bit range getting")
	assert.Equal(t, uint64(0x50), t3.hiz,   "Incorrect bit range getting")
	assert.Equal(t, uint64(0xa0), t3.undef, "Incorrect bit range getting")

	t1.SetBit(0, Undefined)
	t1.SetBit(1, Hi)
	e = val.SetBitRange(16, 23, t1.(*Value64))
	assert.Equal(t, uint64(0x00000a0000), val.bits,  "Incorrect bit range setting")
	assert.Equal(t, uint64(0x1800500000), val.hiz,   "Incorrect bit range setting")
	assert.Equal(t, uint64(0x0000a10000), val.undef, "Incorrect bit range setting")

}

func TestValueUnaryOps(t *testing.T) {
	var val ValueInterface
	sizes := []uint{48, 61, 212, 247}

	for _, size := range sizes {
		val = NewValue(size)

		switch {
		case size <= 64:
			_, ok := val.(*Value64)
			assert.Equal(t, true, ok, "Wrong kind")
		case size > 64:
			_, ok := val.(*ValueBig)
			assert.Equal(t, true, ok, "Wrong kind")
		}

		fmt.Printf("size = %d\n", size)
		val.SetBit(size - 2, Hi)
		inv := val.Unary("~")

		b, _ := inv.GetBit(size - 2)
		assert.Equal(t, Lo, b, "inv incorrect")
	
		or := val.Unary("|")
		b, _ = or.GetBit(0)
		assert.Equal(t, uint(1), or.BitLen(), "or has more than 1 bit")
		assert.Equal(t, Hi, b, "Bit 0 should be Hi")

		nor := val.Unary("~|")
		b, _ = nor.GetBit(0)
		assert.Equal(t, uint(1), nor.BitLen(), "nor has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")
	
		and := val.Unary("&")
		b, _ = and.GetBit(0)
		assert.Equal(t, uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")

		oddParity := val.Unary("^")
		b, _ = oddParity.GetBit(0)
		assert.Equal(t, uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(t, Hi, b, fmt.Sprintf("Bit 0 should be Hi, size = %d", size))
		
		evenParity := val.Unary("~^")
		b, _ = evenParity.GetBit(0)
		assert.Equal(t, uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(t, Lo, b, fmt.Sprintf("Bit 0 should be Lo, size = %d", size))

		
		for i := uint(0); i < val.BitLen(); i++ {
			val.SetBit(i, Hi)
		}
		and = val.Unary("&")
		b, _ = and.GetBit(0)
		assert.Equal(t, uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(t, Hi, b, "Bit 0 should be Hi")

		nand := val.Unary("~&")
		b, _ = nand.GetBit(0)
		assert.Equal(t, uint(1), nand.BitLen(), "nand has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")

		exp := Lo
		notExp := Hi
		if size % 2 != 0 {
			exp = Hi
			notExp = Lo
		}
		oddParity = val.Unary("^")
		b, _ = oddParity.GetBit(0)
		assert.Equal(t, uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(t, exp, b, "Bit 0 should be Lo")
		
		evenParity = val.Unary("~^")
		b, _ = evenParity.GetBit(0)
		assert.Equal(t, uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(t, notExp, b, "Bit 0 should be Hi")

	}

}
