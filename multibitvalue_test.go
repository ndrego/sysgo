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
	assert.Equal(t, uint64(0x1), mbv.bits, "Bit 0 not Hi")
	assert.Equal(t, Hi, mbv.getBit(0), "Bit 0 not Hi")

	mbv.setBit(36, Hi)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 36 not Hi")
	assert.Equal(t, Hi, mbv.getBit(36), "Bit 36 not Hi")

	mbv.setBit(35, HiZ)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x0800000000), mbv.hiz, "Bit 35 not Hiz")
	assert.Equal(t, HiZ, mbv.getBit(35), "Bit 35 not HiZ")

	mbv.setBit(13, Undefined)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 35 not HiZ")
	assert.Equal(t, uint64(0x2000), mbv.undef, "Bit 13 not Undefined")
	assert.Equal(t, Undefined, mbv.getBit(13), "Bit 13 not Undefined")

	mbv.setBit(13, Lo)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x0), mbv.undef, "Bit 13 not Lo")
	assert.Equal(t, uint64(0x800000000), mbv.hiz, "Bit 13 not Lo")
	assert.Equal(t, Lo, mbv.getBit(13), "Bit 13 not Lo")
	
	mbv.setBit(36, HiZ)
	assert.Equal(t, uint64(0x1000000001), mbv.bits, "Bit 36 not Hi")
	assert.Equal(t, uint64(0x1800000000), mbv.hiz, "Bit 36 not HiZ")
	assert.Equal(t, HiZ, mbv.getBit(36), "Bit 36 not HiZ")

	mbv.setBit(0, Lo)
	assert.Equal(t, uint64(0x1000000000), mbv.bits, "Bit 0 not Lo")
	assert.Equal(t, Lo, mbv.getBit(0), "Bit 0 not Lo")
	
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

		assert.Equal(t, Lo, inv.getBit(size - 2), "inv incorrect")
	
		or := mbv.unary("|")
		assert.Equal(t, uint(1), or.bitLen(), "or has more than 1 bit")
		assert.Equal(t, Hi, or.getBit(0), "Bit 0 should be Hi")

		nor := mbv.unary("~|")
		assert.Equal(t, uint(1), nor.bitLen(), "nor has more than 1 bit")
		assert.Equal(t, Lo, nor.getBit(0), "Bit 0 should be Lo")
	
		and := mbv.unary("&")
		assert.Equal(t, uint(1), and.bitLen(), "and has more than 1 bit")
		assert.Equal(t, Lo, and.getBit(0), "Bit 0 should be Lo")

		oddParity := mbv.unary("^")
		assert.Equal(t, uint(1), oddParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, Hi, oddParity.getBit(0), fmt.Sprintf("Bit 0 should be Hi, size = %d", size))
		
		evenParity := mbv.unary("~^")
		assert.Equal(t, uint(1), evenParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, Lo, evenParity.getBit(0), fmt.Sprintf("Bit 0 should be Lo, size = %d", size))

		
		for i := uint(0); i < mbv.bitLen(); i++ {
			mbv.setBit(i, Hi)
		}
		and = mbv.unary("&")
		assert.Equal(t, uint(1), and.bitLen(), "and has more than 1 bit")
		assert.Equal(t, Hi, and.getBit(0), "Bit 0 should be Hi")

		nand := mbv.unary("~&")
		assert.Equal(t, uint(1), nand.bitLen(), "nand has more than 1 bit")
		assert.Equal(t, Lo, nand.getBit(0), "Bit 0 should be Lo")

		exp := Lo
		notExp := Hi
		if size % 2 != 0 {
			exp = Hi
			notExp = Lo
		}
		oddParity = mbv.unary("^")
		assert.Equal(t, uint(1), oddParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, exp, oddParity.getBit(0), "Bit 0 should be Lo")
		
		evenParity = mbv.unary("~^")
		assert.Equal(t, uint(1), evenParity.bitLen(), "parity has more than 1 bit")
		assert.Equal(t, notExp, evenParity.getBit(0), "Bit 0 should be Hi")

	}

}
