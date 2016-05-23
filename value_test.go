package sysgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	_ "reflect"
	"testing"
)

func TestNewValueString(t *testing.T) {
	assert := assert.New(t)
	
	v, e := NewValueString("1'b0")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(1), v.BitLen())
	assert.Equal(Lo, v.GetBit(0))

	v, e = NewValueString("1'b1")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(1), v.BitLen())
	assert.Equal(Hi, v.GetBit(0))
		
	v, e = NewValueString("1'bz")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(1), v.BitLen())
	assert.Equal(HiZ, v.GetBit(0))

	v, e = NewValueString("1'bx")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(1), v.BitLen())
	assert.Equal(Undefined, v.GetBit(0))

	v, e = NewValueString("'b1")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(1), v.BitLen())
	assert.Equal(Hi, v.GetBit(0))

	v, e = NewValueString("32'hdeadbeefe")
	assert.NotNil(e, "Error does not exist but should")

	v, e = NewValueString("32'hdeadbeef")
	assert.Nil(e, "Error exists")
	assert.Equal(uint64(0xdeadbeef), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0),          v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0),          v.(*Value64).undef, "Incorrect value for undef")
	assert.Equal("32'b11011110101011011011111011101111", v.(*Value64).Text('b'), "Incorrect binary text")
	assert.Equal("32'o33653337357", v.(*Value64).Text('o'), "Incorrect octal text")
	assert.Equal("32'hdeadbeef", v.(*Value64).Text('h'), "Incorrect hex text")

	v, e = NewValueString("32'hdeadbezf")
	assert.Nil(e, "Error exists")
	assert.Equal(uint64(0xdeadbe0f), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0xf0),       v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0),          v.(*Value64).undef, "Incorrect value for undef")

	v, e = NewValueString("32'hdxadbezf")
	assert.Nil(e, "Error exists")
	assert.Equal(uint64(0xd0adbe0f), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0x000000f0), v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0x0f000000), v.(*Value64).undef, "Incorrect value for undef")
	assert.Equal("32'b1101xxxx1010110110111110zzzz1111", v.(*Value64).Text('b'), "Incorrect binary text")
	assert.Equal("32'o3xx53337zz7", v.(*Value64).Text('o'), "Incorrect octal text")
	assert.Equal("32'hdxadbezf", v.(*Value64).Text('h'), "Incorrect hex text")
	

	v, e = NewValueString("'hdxadbezf")
	assert.Nil(e, "Error exists")
	assert.Equal(uint(32), v.BitLen(), "Incorrect bit length")
	assert.Equal(uint64(0xd0adbe0f), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0x000000f0), v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0x0f000000), v.(*Value64).undef, "Incorrect value for undef")
	
	v, e = NewValueString("32'h1dxadbezf")
	assert.NotNil(e, "Error should exist")

	// Check octal
	v, e = NewValueString("8'o347")
	assert.Nil(e, "Error exists")
	assert.Equal(uint64(0xe7), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0x00), v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0x00), v.(*Value64).undef, "Incorrect value for undef")

	v, e = NewValueString("7'o347")
	assert.NotNil(e, "Error should exist")

	v, e = NewValueString("8'b1110011x")
	assert.Nil(e, "Error exists")
	assert.Equal(uint64(0xe6), v.(*Value64).bits,  "Incorrect value for bits")
	assert.Equal(uint64(0x00), v.(*Value64).hiz,   "Incorrect value for hiz")
	assert.Equal(uint64(0x01), v.(*Value64).undef, "Incorrect value for undef")

	var exp, zero, hiz, undef big.Int

	v, e = NewValueString("65'h1deadbeef01234567")
	exp.SetString("0x1deadbeef01234567", 0)
	assert.Nil(e, "Error exists")
	assert.Equal(0, v.(*ValueBig).bits.Cmp( &exp),  "Incorrect value for bits")
	assert.Equal(0, v.(*ValueBig).hiz.Cmp(  &zero),  "Incorrect value for hiz")
	assert.Equal(0, v.(*ValueBig).undef.Cmp(&zero), "Incorrect value for undef")
	assert.Equal("65'h1deadbeef01234567", v.(*ValueBig).Text('h'))

	v, e = NewValueString("68'hzdeadbeef01234567")
	exp.SetString("0x0deadbeef01234567", 0)
	hiz.SetString("0xf0000000000000000", 0)
	assert.Nil(e, "Error exists")
	assert.Equal(0, v.(*ValueBig).bits.Cmp( &exp),  "Incorrect value for bits")
	assert.Equal(0, v.(*ValueBig).hiz.Cmp(  &hiz),  "Incorrect value for hiz")
	assert.Equal(0, v.(*ValueBig).undef.Cmp(&zero), "Incorrect value for undef")

	v, e = NewValueString("65'hzdeadzxef01234567")
	exp.SetString(  "0x0dead00ef01234567", 0)
	hiz.SetString(  "0x10000f00000000000", 0)
	undef.SetString("0x000000f0000000000", 0)
	assert.Nil(e, "Error exists")
	assert.Equal(0, v.(*ValueBig).bits.Cmp( &exp),   "Incorrect value for bits")
	assert.Equal(0, v.(*ValueBig).hiz.Cmp(  &hiz),   "Incorrect value for hiz")
	assert.Equal(0, v.(*ValueBig).undef.Cmp(&undef), "Incorrect value for undef")
	assert.Equal("65'hzdeadzxef01234567", v.(*ValueBig).Text('h'))	
}

func TestValue1GetSetBit(t *testing.T) {
	assert := assert.New(t)
	
	v1, ok1 := NewValue(1).(*Value1)
	assert.Equal(true, ok1, "val not a Value1")

	assert.NotNil(v1.SetBit(1, Hi), "Should have error, but have nil")

	states := []LogicState{Lo, Hi, Undefined, HiZ}
	for _, s := range states {
		v1.SetBit(0, s)
		b := v1.GetBit(0)
		assert.Equal(s, b, fmt.Sprintf("bit should be %s but is %s", s, b))
		assert.Equal(s, v1.v, fmt.Sprintf("bit should be %s but is %s", s, b))
	}

	// Make sure the range functions for Value1 return errors
	assert.Nil(v1.GetBits([]uint{0}), "Should have nil from Value1.GetBits()")
	assert.Nil(v1.GetBitRange(0, 31), "Should have nil from Value1.GetBitRange()")
	assert.NotNil(v1.SetBitRange(0, 31, NewValue(32)), "Should have error from Value1.SetBitRange()")
}

func TestValue64GetSetBit(t *testing.T) {
	assert := assert.New(t)

	val, ok := NewValue(37).(*Value64)
	assert.Equal(true, ok, "val not a Value64")
	
	// Set lo/hi first
	val.SetBit(0, Hi)
	assert.Equal(uint64(0x1), val.bits, "Bit 0 not Hi")
	assert.Equal(Hi, val.GetBit(0), "Bit 0 not Hi")

	val.SetBit(36, Hi)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 36 not Hi")
	assert.Equal(Hi, val.GetBit(36), "Bit 36 not Hi")

	val.SetBit(35, HiZ)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(uint64(0x0800000000), val.hiz, "Bit 35 not Hiz")
	assert.Equal(HiZ, val.GetBit(35), "Bit 35 not HiZ")

	val.SetBit(13, Undefined)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 35 not HiZ")
	assert.Equal(uint64(0x2000), val.undef, "Bit 13 not Undefined")
	assert.Equal(Undefined, val.GetBit(13), "Bit 13 not Undefined")

	val.SetBit(13, Lo)
	assert.Equal(uint64(0x1000000001), val.bits, "Bit 13 not Lo")
	assert.Equal(uint64(0x0), val.undef, "Bit 13 not Lo")
	assert.Equal(uint64(0x800000000), val.hiz, "Bit 13 not Lo")
	assert.Equal(Lo, val.GetBit(13), "Bit 13 not Lo")
	
	val.SetBit(36, HiZ)
	assert.Equal(uint64(0x0000000001), val.bits, "Bit 36 not HiZ")
	assert.Equal(uint64(0x1800000000), val.hiz, "Bit 36 not HiZ")
	assert.Equal(HiZ, val.GetBit(36), "Bit 36 not HiZ")

	val.SetBit(0, Lo)
	assert.Equal(uint64(0x0000000000), val.bits, "Bit 0 not Lo")
	assert.Equal(Lo, val.GetBit(0), "Bit 0 not Lo")

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

	t2 := val.GetBitRange(0, 38)
	assert.Nil(t2, "Should have nil from GetBitRange")
	t2 = val.GetBitRange(16, 23)
	t3 := t2.(*Value64)
	assert.Equal(uint64(0x09), t3.bits,  "Incorrect bit range getting")
	assert.Equal(uint64(0x50), t3.hiz,   "Incorrect bit range getting")
	assert.Equal(uint64(0xa0), t3.undef, "Incorrect bit range getting")

	t1.SetBit(0, Undefined)
	t1.SetBit(1, Hi)
	val.SetBitRange(16, 23, t1.(*Value64))
	assert.Equal(uint64(0x00000a0000), val.bits,  "Incorrect bit range setting")
	assert.Equal(uint64(0x1800500000), val.hiz,   "Incorrect bit range setting")
	assert.Equal(uint64(0x0000a10000), val.undef, "Incorrect bit range setting")

}

func TestValueConcat(t *testing.T) {
	assert := assert.New(t)

	x, _ := NewValueString("1'b1")
	y, _ := NewValueString("1'b0")
	z := x.Concat(y)
	assert.Equal("2'b10", z.Text('b'), "Incorrect concatenation")

	y, _ = NewValueString("1'bx")
	z = x.Concat(y)
	assert.Equal("2'b1x", z.Text('b'), "Incorrect concatenation")

	x, _ = NewValueString("1'bz")
	z = x.Concat(y)
	assert.Equal("2'bzx", z.Text('b'), "Incorrect concatenation")

	x, _ = NewValueString("3'b101")
	y, _ = NewValueString("1'b0")
	z = x.Concat(y)
	assert.Equal("4'b1010", z.Text('b'), "Incorrect concatenation")

	z = y.Concat(x)
	assert.Equal("4'b0101", z.Text('b'), "Incorrect concatenation")
	
	x, _ = NewValueString("65'h1deadbeef01234567")
	y, _ = NewValueString("3'b101")
	z = y.Concat(x)
	assert.Equal("68'hbdeadbeef01234567", z.Text('h'), "Incorrect concatenation")

	z = x.Concat(y)
	assert.Equal("68'hef56df778091a2b3d", z.Text('h'), "Incorrect concatenation")

	y, _ = NewValueString("1'b1")
	z = x.Concat(y)
	assert.Equal("66'h3bd5b7dde02468acf", z.Text('h'), "Incorrect concatenation")

	z = y.Concat(x)
	assert.Equal("66'h3deadbeef01234567", z.Text('h'), "Incorrect concatenation")

	y, _ = NewValueString("32'hbeefdead")
	z = x.Concat(y)
	assert.Equal("97'h1deadbeef01234567beefdead", z.Text('h'), "Incorrect concatenation")

	z = y.Concat(x)
	assert.Equal("97'h17ddfbd5bdeadbeef01234567", z.Text('h'), "Incorrect concatenation")	
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
		assert.Equal(exp, inv.GetBit(0), "Incorrect inversion for %s", s)
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

		assert.Equal(Lo, inv.GetBit(size - 2), "inv incorrect")
	
		or := val.Unary("|")
		assert.Equal(uint(1), or.BitLen(), "or has more than 1 bit")
		assert.Equal(Hi, or.GetBit(0), "Bit 0 should be Hi")

		nor := val.Unary("~|")
		assert.Equal(uint(1), nor.BitLen(), "nor has more than 1 bit")
		assert.Equal(Lo, nor.GetBit(0), "Bit 0 should be Lo")
	
		and := val.Unary("&")
		assert.Equal(uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(Lo, and.GetBit(0), "Bit 0 should be Lo")

		oddParity := val.Unary("^")
		assert.Equal(uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(Hi, oddParity.GetBit(0), fmt.Sprintf("Bit 0 should be Hi, size = %d", size))
		
		evenParity := val.Unary("~^")
		assert.Equal(uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(Lo, evenParity.GetBit(0), fmt.Sprintf("Bit 0 should be Lo, size = %d", size))

		
		for i := uint(0); i < val.BitLen(); i++ {
			val.SetBit(i, Hi)
		}
		and = val.Unary("&")
		assert.Equal(uint(1), and.BitLen(), "and has more than 1 bit")
		assert.Equal(Hi, and.GetBit(0), "Bit 0 should be Hi")

		nand := val.Unary("~&")
		assert.Equal(uint(1), nand.BitLen(), "nand has more than 1 bit")
		assert.Equal(Lo, nand.GetBit(0), "Bit 0 should be Lo")

		exp := Lo
		notExp := Hi
		if size % 2 != 0 {
			exp = Hi
			notExp = Lo
		}
		oddParity = val.Unary("^")
		assert.Equal(uint(1), oddParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(exp, oddParity.GetBit(0), "Bit 0 should be Lo")
		
		evenParity = val.Unary("~^")
		assert.Equal(uint(1), evenParity.BitLen(), "parity has more than 1 bit")
		assert.Equal(notExp, evenParity.GetBit(0), "Bit 0 should be Hi")

	}

}
