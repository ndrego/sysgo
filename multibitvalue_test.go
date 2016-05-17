package sysgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	_ "math"
	_ "reflect"
	"testing"
)

func TestMBVGetSetBit(t *testing.T) {
	mbv, ok := NewMultiBitValue(37).(*mbv64)
	assert.Equal(t, true, ok, "mbv not an mbv64")
	
	// Set lo/hi first
	mbv.setBit(0, Hi)
	b, _ := mbv.getBit(0)
	assert.Equal(t, uint64(0x1), mbv.bits, "Bit 0 not Hi")
	assert.Equal(t, Hi, b, "Bit 0 not Hi")

	mbv.setBit(36, Hi)
	b, _ = mbv.getBit(36)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 36 not Hi")
	assert.Equal(t, Hi, b, "Bit 36 not Hi")

	mbv.setBit(35, HiZ)
	b, _ = mbv.getBit(35)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x0800000000), mbv.hiz, "Bit 35 not Hiz")
	assert.Equal(t, HiZ, b, "Bit 35 not HiZ")

	mbv.setBit(13, Undefined)
	b, _ = mbv.getBit(13)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x2000), mbv.undef, "Bit 13 not Undefined")
	assert.Equal(t, Undefined, b, "Bit 13 not Undefined")

	mbv.setBit(13, Lo)
	b, _ = mbv.getBit(13)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x0), mbv.undef, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x800000000), mbv.hiz, "Bit 13 not Lo")
	assert.Equal(t, Lo, b, "Bit 13 not Lo")
	
	mbv.setBit(36, HiZ)
	b, _ = mbv.getBit(36)
	assert.Equal(t, uint64(0x0000000001), mbv.bits, "Bit 36 not HiZ")
	assert.Equal(t, uint64(0x1800000000), mbv.hiz, "Bit 36 not HiZ")
	assert.Equal(t, HiZ, b, "Bit 36 not HiZ")

	mbv.setBit(0, Lo)
	b, _ = mbv.getBit(0)
	assert.Equal(t, uint64(0x0000000000), mbv.bits, "Bit 0 not Lo")
	assert.Equal(t, Lo, b, "Bit 0 not Lo")

	t1 := NewMultiBitValue(8)
	t1.setBit(0, Hi)
	t1.setBit(1, Lo)
	t1.setBit(2, Lo)
	t1.setBit(3, Hi)
	t1.setBit(4, HiZ)
	t1.setBit(5, Undefined)
	t1.setBit(6, HiZ)
	t1.setBit(7, Undefined)
	e := mbv.setBitRange(16, 23, t1.(*mbv64))
	assert.Equal(t, nil, e, "Error when setting bit range")

	assert.Equal(t, uint64(0x0000090000), mbv.bits,  "Incorrect bit range setting")
	assert.Equal(t, uint64(0x1800500000), mbv.hiz,   "Incorrect bit range setting")
	assert.Equal(t, uint64(0x0000a00000), mbv.undef, "Incorrect bit range setting")

	t2, e := mbv.getBitRange(0, 38)
	assert.Equal(t, true, e != nil, "Should have error in getBitRange, but none received")
	t2, _ = mbv.getBitRange(16, 23)
	t3 := t2.(*mbv64)
	assert.Equal(t, uint64(0x09), t3.bits,  "Incorrect bit range getting")
	assert.Equal(t, uint64(0x50), t3.hiz,   "Incorrect bit range getting")
	assert.Equal(t, uint64(0xa0), t3.undef, "Incorrect bit range getting")

	t1.setBit(0, Undefined)
	t1.setBit(1, Hi)
	e = mbv.setBitRange(16, 23, t1.(*mbv64))
	assert.Equal(t, uint64(0x00000a0000), mbv.bits,  "Incorrect bit range setting")
	assert.Equal(t, uint64(0x1800500000), mbv.hiz,   "Incorrect bit range setting")
	assert.Equal(t, uint64(0x0000a10000), mbv.undef, "Incorrect bit range setting")

}

func TestMBVUnaryOps(t *testing.T) {
	var mbv MultiBitValue
	sizes := []uint{48, 61, 212, 247}

	for _, size := range sizes {
		mbv = NewMultiBitValue(size)

		switch {
		case size <= 64:
			_, ok := mbv.(*mbv64)
			assert.Equal(t, true, ok, "Wrong kind")
		case size > 64:
			_, ok := mbv.(*mbvBig)
			assert.Equal(t, true, ok, "Wrong kind")
		}

		fmt.Printf("size = %d\n", size)
		mbv.setBit(size - 2, Hi)
		inv := mbv.unary("~")

		b, _ := inv.getBit(size - 2)
		assert.Equal(t, Lo, b, "inv incorrect")
	
		or := mbv.unary("|")
		b, _ = or.getBit(0)
		assert.Equal(t, uint(1), or.bitLen(), "or has more than 1 bit")
		assert.Equal(t, Hi, b, "Bit 0 should be Hi")

		nor := mbv.unary("~|")
		b, _ = nor.getBit(0)
		assert.Equal(t, uint(1), nor.bitLen(), "nor has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")
	
		and := mbv.unary("&")
		b, _ = and.getBit(0)
		assert.Equal(t, uint(1), and.bitLen(), "and has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")

		oddParity := mbv.unary("^")
		b, _ = oddParity.getBit(0)
		assert.Equal(t, uint(1), oddParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, Hi, b, fmt.Sprintf("Bit 0 should be Hi, size = %d", size))
		
		evenParity := mbv.unary("~^")
		b, _ = evenParity.getBit(0)
		assert.Equal(t, uint(1), evenParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, Lo, b, fmt.Sprintf("Bit 0 should be Lo, size = %d", size))

		
		for i := uint(0); i < mbv.bitLen(); i++ {
			mbv.setBit(i, Hi)
		}
		and = mbv.unary("&")
		b, _ = and.getBit(0)
		assert.Equal(t, uint(1), and.bitLen(), "and has more than 1 bit")
		assert.Equal(t, Hi, b, "Bit 0 should be Hi")

		nand := mbv.unary("~&")
		b, _ = nand.getBit(0)
		assert.Equal(t, uint(1), nand.bitLen(), "nand has more than 1 bit")
		assert.Equal(t, Lo, b, "Bit 0 should be Lo")

		exp := Lo
		notExp := Hi
		if size % 2 != 0 {
			exp = Hi
			notExp = Lo
		}
		oddParity = mbv.unary("^")
		b, _ = oddParity.getBit(0)
		assert.Equal(t, uint(1), oddParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, exp, b, "Bit 0 should be Lo")
		
		evenParity = mbv.unary("~^")
		b, _ = evenParity.getBit(0)
		assert.Equal(t, uint(1), evenParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, notExp, b, "Bit 0 should be Hi")

	}

}
