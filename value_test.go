package sysgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	_ "reflect"
	"strings"
	"testing"
)

type binaryTest struct {
	exp LogicState
	op1 ValueInterface
	op string
	op2 ValueInterface
}

func runBinaryTests(t *testing.T, tests []binaryTest) {
	assert := assert.New(t)
	for _, test := range tests {
		assert.Equal(test.exp, test.op1.Binary(test.op, test.op2).GetBit(0), "Exp = %s, %s %s %s\n", test.exp, test.op1, test.op, test.op2)
	}
}

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

func TestValueLsh(t *testing.T) {
	assert := assert.New(t)

	v1, _ := NewValueString("1'b1")
	v2, _ := NewValueString("1'b0")
	v3, _ := NewValueString("1'bz")
	v4, _ := NewValueString("32'hdeadbeef")
	v5, _ := NewValueString("32'hdxadbeef")
	v6, _ := NewValueString("128'hdeadbeef")

	assert.Equal("1'b1",                  v1.Lsh(0).Text( 'b'))
	assert.Equal("2'b10",                 v1.Lsh(1).Text( 'b'))
	assert.Equal("65'h10000000000000000", v1.Lsh(64).Text('h'))

	assert.Equal("1'b0",                  v2.Lsh(0).Text( 'b'))
	assert.Equal("2'b00",                 v2.Lsh(1).Text( 'b'))
	assert.Equal("65'h00000000000000000", v2.Lsh(64).Text('h'))

	assert.Equal("1'bz",                  v3.Lsh(0).Text( 'b'))
	assert.Equal("2'bz0",                 v3.Lsh(1).Text( 'b'))
	assert.Equal("65'hz0000000000000000", v3.Lsh(64).Text('h'))

	assert.Equal("32'hdeadbeef",          v4.Lsh(0).Text( 'h'))
	assert.Equal("33'h1bd5b7dde",         v4.Lsh(1).Text( 'h'))
	assert.Equal("65'h1bd5b7dde00000000", v4.Lsh(33).Text('h'))

	assert.Equal("32'hdxadbeef",          v5.Lsh(0).Text( 'h'))
	assert.Equal("33'h1xx5b7dde",         v5.Lsh(1).Text( 'h'))
	assert.Equal("65'h1xx5b7dde00000000", v5.Lsh(33).Text('h'))

	assert.Equal("128'h000000000000000000000000deadbeef",          v6.Lsh(0).Text( 'h'))
	assert.Equal("129'h0000000000000000000000001bd5b7dde",         v6.Lsh(1).Text( 'h'))
	assert.Equal("161'h0000000000000000000000001bd5b7dde00000000", v6.Lsh(33).Text('h'))
}

func TestValueRsh(t *testing.T) {
	assert := assert.New(t)

	v1, _ := NewValueString("1'b1")
	v2, _ := NewValueString("1'b0")
	v3, _ := NewValueString("1'bz")
	v4, _ := NewValueString("32'hdeadbeef")
	v5, _ := NewValueString("32'hdxadbeef")
	v6, _ := NewValueString("128'hdeadbeef0000000000000000beefdead")

	assert.Equal("1'b1", v1.Rsh(0).Text( 'b'))
	assert.Equal("1'b0", v1.Rsh(1).Text( 'b'))
	assert.Equal("1'b0", v1.Rsh(64).Text('b'))

	assert.Equal("1'b0", v2.Rsh(0).Text( 'b'))
	assert.Equal("1'b0", v2.Rsh(1).Text( 'b'))
	assert.Equal("1'b0", v2.Rsh(64).Text('b'))

	assert.Equal("1'bz", v3.Rsh(0).Text( 'b'))
	assert.Equal("1'b0", v3.Rsh(1).Text( 'b'))
	assert.Equal("1'b0", v3.Rsh(64).Text('b'))

	assert.Equal("32'hdeadbeef", v4.Rsh(0).Text( 'h'))
	assert.Equal("32'h6f56df77", v4.Rsh(1).Text( 'h'))
	assert.Equal("32'h00000001", v4.Rsh(31).Text('h'))
	assert.Equal("32'h00000000", v4.Rsh(33).Text('h'))

	assert.Equal("32'hdxadbeef", v5.Rsh(0).Text( 'h'))
	assert.Equal("32'h6xx6df77", v5.Rsh(1).Text( 'h'))
	assert.Equal("32'h00000001", v5.Rsh(31).Text('h'))
	assert.Equal("32'h00000000", v5.Rsh(33).Text('h'))	

	assert.Equal("128'hdeadbeef0000000000000000beefdead", v6.Rsh(0).Text(  'h'))
	assert.Equal("128'h6f56df7780000000000000005f77ef56", v6.Rsh(1).Text(  'h'))
	assert.Equal("128'h0000000000000000deadbeef00000000", v6.Rsh(64).Text( 'h'))
	assert.Equal("128'h00000000000000000000000000000000", v6.Rsh(129).Text('h'))
}

func TestValueCmpEquality(t *testing.T) {
	a := NewValue(1)
	b, _ := NewValueString("1'b1")
	c, _ := NewValueString("1'b0")
	d, _ := NewValueString("1'bx")
	e, _ := NewValueString("1'bz")
	f, _ := NewValueString("4'b1010")
	g, _ := NewValueString("4'b101z")
	h, _ := NewValueString("4'b0001")
	i, _ := NewValueString("4'b101x")
	j, _ := NewValueString("5'b0101x")
	k, _ := NewValueString("72'hdeadbeef")
	l, _ := NewValueString("66'hdeadbeef")
	m, _ := NewValueString("32'hdeadbeef")
	n, _ := NewValueString("64'h1")
	o, _ := NewValueString("64'b101z")

	k = k.Lsh(32)
	l = l.Lsh(32)

	runBinaryTests(t, []binaryTest{
		{Hi,        a, "==",  c},
		{Lo,        a, "==",  b},
		{Hi,        a, "!=",  b},
		{Lo,        a, "!=",  c},
		{Undefined, a, "==",  d},
		{Undefined, a, "!=",  d},
		{Undefined, d, "==",  e},
		{Lo,        d, "===", e},
		{Hi,        d, "!==", e},
		{Hi,        b, "==",  h},
		{Lo,        c, "==",  h},
		{Hi,        c, "!=",  h},
		{Hi,        f, "==",  f},
		{Undefined, f, "==",  g},
		{Lo,        f, "!=",  f},
		{Undefined, f, "!=",  g},
		{Undefined, g, "==",  i},
		{Undefined, i, "==",  g},
		{Undefined, g, "!=",  i},
		{Lo,        i, "===", g},
		{Hi,        i, "!==", g},
		{Hi,        i, "===", j},
		{Hi,        k, "==",  l},
		{Hi,        l, "==",  k},
		{Lo,        k, "!=",  l},
		{Hi,        k, "!=",  f},
		{Undefined, k, "!=",  g},
		{Hi,        m, "==",  k.Rsh(32)},
		{Hi,        n, "==",  b},
		{Undefined, o, "==",  g},
		{Undefined, g, "==",  o},
		{Hi,        o, "===", g},
		{Lo,        o, "!==", g},
		{Hi,        g, "===", o},
		{Lo,        g, "!==", o},
		{Hi,        i, "!==", o},
		{Lo,        i, "===", o},
	})

}

func TestValueCmpRelational(t *testing.T) {
	a := NewValue(1)
	b, _ := NewValueString("1'b1")
	c, _ := NewValueString("1'b0")
	d, _ := NewValueString("1'bx")
	e, _ := NewValueString("1'bz")
	f, _ := NewValueString("4'b1010")
	g, _ := NewValueString("4'b101z")
	h, _ := NewValueString("4'b0001")
	i, _ := NewValueString("4'b101x")
	j, _ := NewValueString("5'b0101x")
	k, _ := NewValueString("72'hdeadbeef")
	l, _ := NewValueString("66'hdeadbeef")
	m, _ := NewValueString("32'hdeadbeef")
	n, _ := NewValueString("64'h1")
	o, _ := NewValueString("64'b101z")


	tests := []binaryTest {
		{Hi, a, "<=", b},
		{Hi, a, "<",  b},
		{Lo, a, ">=", b},
		{Lo, a, ">",  b},
		{Hi, a, "<=", c},
		{Lo, a, "<",  c},
		{Hi, c, "<=", a},
		{Lo, c, "<",  a},
		{Hi, b, "<=", f},
		{Hi, b, "<",  f},
		{Lo, b, ">",  f},
		{Lo, b, ">=", f},
		{Lo, f, "<=", b},
		{Lo, f, "<",  b},
		{Hi, f, ">",  b},
		{Hi, f, ">=", b},
		{Hi, f, ">",  h},
		{Hi, f, ">=", h},
		{Lo, f, "<",  h},
		{Lo, f, "<=", h},
		{Lo, h, ">",  f},
		{Lo, h, ">=", f},
		{Hi, h, "<",  f},
		{Hi, h, "<=", f},
		{Lo, k, ">",  l},
		{Lo, k, ">",  m},
		{Hi, k, ">=", l},
		{Hi, k, ">=", m},
		{Hi, k, "<=", l},
		{Hi, k, "<=", m},
		{Lo, k, "<",  l},
		{Lo, k, "<",  m},
	}
		
	ops := []string{"<", "<=", ">", ">="}
	for _, op := range ops {	
		for _, x := range []ValueInterface{a, b, c} {
			y := []binaryTest {
				{Undefined, x, op, d},
				{Undefined, x, op, e},
				{Undefined, d, op, x},
				{Undefined, e, op, x},
			}
			tests = append(tests, y...)
		}
		for _, x := range []ValueInterface{g, i, j, o} {
			y := []binaryTest {
				{Undefined, f, op, x},
				{Undefined, h, op, x},
				{Undefined, x, op, f},
				{Undefined, x, op, h},
				{Undefined, x, op, x},
				{Undefined, x, op, x},
			}
			tests = append(tests, y...)
		}

		var exp LogicState
		for _, x := range []ValueInterface{b, h} {
			if strings.HasSuffix(op, "=") {
				exp = Hi
			} else {
				exp = Lo
			}
			y := []binaryTest {
				{exp, n, op, x},
			}
			tests = append(tests, y...)
		}
		for _, x := range []ValueInterface{f, k, l, m} {
			if strings.HasPrefix(op, "<") {
				exp = Hi
			} else {
				exp = Lo
			}
			y := []binaryTest {
				{exp, n, op, x},
				{exp.Unary('~'), x, op, n},
			}
			tests = append(tests, y...)
		}
	}

	runBinaryTests(t, tests)
}

func TestValueBinLogical(t *testing.T) {
	a := NewValue(1)
	b, _ := NewValueString("1'b1")
	c, _ := NewValueString("1'b0")
	d, _ := NewValueString("1'bx")
	e, _ := NewValueString("1'bz")
	f, _ := NewValueString("4'b1010")
	g, _ := NewValueString("4'b101z")
	h, _ := NewValueString("4'b0001")
	i, _ := NewValueString("4'b101x")
	j, _ := NewValueString("5'b0101x")
	k, _ := NewValueString("4'b0")
	l, _ := NewValueString("66'hdeadbeef")
	m, _ := NewValueString("32'hdeadbeef")
	n, _ := NewValueString("65'h0")
	o, _ := NewValueString("64'b101z")
	p, _ := NewValueString("65'bz")

	tests := []binaryTest {
		{Hi,        a, "||", b},
		{Lo,        a, "||", a},
		{Lo,        a, "||", c},
		{Undefined, a, "||", d},
		{Hi,        b, "||", d},
		{Undefined, d, "||", e},
		{Lo,        a, "&&", b},
		{Lo,        a, "&&", a},
		{Lo,        a, "&&", c},
		{Undefined, a, "&&", d},
		{Undefined, b, "&&", d},
		{Undefined, d, "&&", e},
		{Hi,        b, "&&", b},

		{Hi,        f, "||", g},
		{Hi,        f, "||", h},
		{Hi,        f, "||", j},
		{Hi,        f, "||", a},
		{Hi,        h, "||", a},
		{Hi,        h, "||", j},
		{Hi,        i, "||", j},
		{Lo,        c, "||", k},
		{Undefined, k, "||", d},
		{Hi,        f, "&&", g},
		{Hi,        f, "&&", h},
		{Lo,        f, "&&", k},
		{Lo,        g, "&&", k},
		{Undefined, f, "&&", d},
		{Undefined, k, "&&", d},
		{Lo,        b, "&&", k},
		{Lo,        c, "&&", k},

		{Hi,        l, "||", m},
		{Hi,        l, "||", n},
		{Hi,        l, "||", f},
		{Hi,        l, "||", k},
		{Hi,        l, "||", a},
		{Hi,        l, "||", b},
		{Hi,        l, "||", d},
		{Hi,        o, "||", d},
		{Lo,        c, "||", n},
		{Lo,        k, "||", n},
		{Undefined, c, "||", p},
		{Undefined, e, "||", p},
		{Hi,        l, "&&", m},
		{Hi,        l, "&&", o},
		{Undefined, l, "&&", p},
		{Lo,        l, "&&", n},
		{Lo,        m, "&&", n},
		{Lo,        o, "&&", n},
	}
	runBinaryTests(t, tests)
}

func TestValueBinBitwise(t *testing.T) {
	assert := assert.New(t)
	
	a := NewValue(1)
	b, _ := NewValueString("1'b1")
	c, _ := NewValueString("1'b0")
	d, _ := NewValueString("1'bx")
	e, _ := NewValueString("1'bz")
	f, _ := NewValueString("4'b1010")
	g, _ := NewValueString("4'b101z")
	h, _ := NewValueString("4'b0001")
	i, _ := NewValueString("4'b101x")
	j, _ := NewValueString("5'b0101x")
	k, _ := NewValueString("4'b0")
	l, _ := NewValueString("66'hdeadbeef")
	m, _ := NewValueString("32'hdeadbeef")
	n, _ := NewValueString("65'h0")
	o, _ := NewValueString("64'b101z")
	p, _ := NewValueString("65'bz")

	assert.Equal(Hi, a.Binary("&",  b).Binary("===", a).GetBit(0))
	assert.Equal(Hi, a.Binary("&",  d).Binary("===", d).GetBit(0))
	assert.Equal(Hi, a.Binary("&",  e).Binary("===", d).GetBit(0))
	assert.Equal(Hi, c.Binary("&",  b).Binary("===", a).GetBit(0))
	assert.Equal(Hi, b.Binary("&",  b).Binary("===", b).GetBit(0))
	assert.Equal(Hi, a.Binary("|",  b).Binary("===", b).GetBit(0))
	assert.Equal(Hi, a.Binary("|",  c).Binary("===", c).GetBit(0))
	assert.Equal(Hi, b.Binary("|",  d).Binary("===", b).GetBit(0))
	assert.Equal(Hi, b.Binary("|",  e).Binary("===", b).GetBit(0))
	assert.Equal(Hi, b.Binary("^",  e).Binary("===", d).GetBit(0))
	assert.Equal(Hi, a.Binary("^",  c).Binary("===", a).GetBit(0))
	assert.Equal(Hi, b.Binary("^",  c).Binary("===", b).GetBit(0))
	assert.Equal(Hi, b.Binary("^",  b).Binary("===", c).GetBit(0))
	assert.Equal(Hi, b.Binary("^~", e).Binary("===", d).GetBit(0))
	assert.Equal(Hi, a.Binary("^~", c).Binary("===", b).GetBit(0))
	assert.Equal(Hi, b.Binary("^~", c).Binary("===", c).GetBit(0))
	assert.Equal(Hi, b.Binary("^~", b).Binary("===", b).GetBit(0))

	fAndg, _ := NewValueString("4'b1010")
	fAndh, _ := NewValueString("4'b0")
	gAndh, _ := NewValueString("4'b000x")
	fAndi := fAndg
	fOrh,  _ := NewValueString("4'b1011")
	iXorj, _ := NewValueString("5'b0000x")
	fXnorg, _ := NewValueString("4'b111x")
	gXnorh, _ := NewValueString("4'b010x")
	fXnorh, _ := NewValueString("4'b0100")
	bXnorh, _ := NewValueString("4'hf")
	assert.Equal(Hi, f.Binary("&",  g).Binary("===", fAndg).GetBit(0))
	assert.Equal(Hi, f.Binary("&",  h).Binary("===", fAndh).GetBit(0))
	assert.Equal(Hi, g.Binary("&",  h).Binary("===", gAndh).GetBit(0))
	assert.Equal(Hi, g.Binary("&",  i).Binary("===", i).GetBit(0))
	assert.Equal(Hi, f.Binary("&",  i).Binary("===", fAndi).GetBit(0))
	assert.Equal(Hi, h.Binary("&",  b).Binary("===", h).GetBit(0))
	assert.Equal(Hi, h.Binary("&",  c).Binary("===", k).GetBit(0))
	assert.Equal(Hi, f.Binary("|",  g).Binary("===", i).GetBit(0))
	assert.Equal(Hi, f.Binary("|",  h).Binary("===", fOrh).GetBit(0))
	assert.Equal(Hi, g.Binary("|",  d).Binary("===", i).GetBit(0))
	assert.Equal(Hi, d.Binary("|",  g).Binary("===", i).GetBit(0))
	assert.Equal(Hi, i.Binary("|",  j).Binary("===", j).GetBit(0))
	assert.Equal(Hi, j.Binary("|",  i).Binary("===", j).GetBit(0))
	assert.Equal(Hi, j.Binary("^",  i).Binary("===", iXorj).GetBit(0))
	assert.Equal(Hi, i.Binary("^",  j).Binary("===", iXorj).GetBit(0))
	assert.Equal(Hi, i.Binary("^",  h).Binary("===", i).GetBit(0))
	assert.Equal(Hi, k.Binary("^",  k).Binary("===", k).GetBit(0))
	assert.Equal(Hi, f.Binary("^",  h).Binary("===", fOrh).GetBit(0))
	assert.Equal(Hi, g.Binary("^",  h).Binary("===", i).GetBit(0))
	assert.Equal(Hi, g.Binary("^~", f).Binary("===", fXnorg).GetBit(0))
	assert.Equal(Hi, g.Binary("^~", h).Binary("===", gXnorh).GetBit(0))
	assert.Equal(Hi, f.Binary("^~", h).Binary("===", fXnorh).GetBit(0))
	assert.Equal(Hi, b.Binary("^~", h).Binary("===", bXnorh).GetBit(0))

	lAndn, _ := NewValueString("65'h0")
	lAndo, _ := NewValueString("66'b101x")
	oOrg,  _ := NewValueString("64'b101x")
	oOrh,  _ := NewValueString("64'b1011")
	lXoro, _ := NewValueString("66'hdeadbee4")
	lXoro.SetBit(0, Undefined)
	lXnorm, _ := NewValueString("66'h3ffffffffffffffff")
	lXnoro, _ := NewValueString("66'h3ffffffff2152411a")
	lXnoro.SetBit(0, Undefined)
	eXnorp, _ := NewValueString("65'h1ffffffffffffffff")
	eXnorp.SetBit(0, Undefined)
	assert.Equal(Hi, l.Binary("&",  m).Binary("===", l).GetBit(0))
	assert.Equal(Hi, l.Binary("&",  n).Binary("===", lAndn).GetBit(0))
	assert.Equal(Hi, l.Binary("&",  n).Binary("!==", l).GetBit(0))
	assert.Equal(Hi, l.Binary("&",  o).Binary("===", lAndo).GetBit(0))
	assert.Equal(Hi, o.Binary("&",  g).Binary("===", lAndo).GetBit(0))	
	assert.Equal(Hi, n.Binary("&",  b).Binary("===", n).GetBit(0))
	assert.Equal(Hi, n.Binary("&",  c).Binary("===", n).GetBit(0))
	assert.Equal(Hi, b.Binary("&",  n).Binary("===", n).GetBit(0))
	assert.Equal(Hi, c.Binary("&",  n).Binary("===", n).GetBit(0))
	assert.Equal(Hi, p.Binary("&",  e).Binary("===", d).GetBit(0))
	assert.Equal(Hi, l.Binary("|",  m).Binary("===", l).GetBit(0))
	assert.Equal(Hi, l.Binary("|",  n).Binary("===", l).GetBit(0))
	assert.Equal(Hi, l.Binary("|",  o).Binary("===", l).GetBit(0))
	assert.Equal(Hi, o.Binary("|",  g).Binary("===", oOrg).GetBit(0))
	assert.Equal(Hi, n.Binary("|",  b).Binary("===", b).GetBit(0))
	assert.Equal(Hi, n.Binary("|",  c).Binary("===", c).GetBit(0))
	assert.Equal(Hi, o.Binary("|",  h).Binary("===", oOrh).GetBit(0))
	assert.Equal(Hi, l.Binary("^",  m).Binary("===", a).GetBit(0))
	assert.Equal(Hi, l.Binary("^",  o).Binary("===", lXoro).GetBit(0))
	assert.Equal(Hi, l.Binary("^",  g).Binary("===", lXoro).GetBit(0))
	assert.Equal(Hi, e.Binary("^",  p).Binary("===", d).GetBit(0))
	assert.Equal(Hi, e.Binary("^",  n).Binary("===", d).GetBit(0))
	assert.Equal(Hi, l.Binary("^~", m).Binary("===", lXnorm).GetBit(0))
	assert.Equal(Hi, l.Binary("^~", o).Binary("===", lXnoro).GetBit(0))
	assert.Equal(Hi, l.Binary("^~", g).Binary("===", lXnoro).GetBit(0))
	assert.Equal(Hi, e.Binary("^~", p).Binary("===", eXnorp).GetBit(0))
	assert.Equal(Hi, e.Binary("^~", n).Binary("===", eXnorp).GetBit(0))
	
}
