package sysgo

import (
)

//go:generate stringer -type Value

type Value uint8

const (
	Lo Value = iota
	Hi
	HiZ
	Undefined
)


func CombineValue(cur, next Value) Value {
	if cur == next {
		return cur
	} else if cur == Undefined || next == Undefined {
		return Undefined
	} else if cur == HiZ && (next == Lo || next == Hi) {
		return next
	} else if (cur == Lo || cur == Hi) && next == HiZ {
		return cur
	}

	return Undefined
}

func (A Value) Invert() Value {
	switch {
	case A == Lo:
		return Hi
	case A == Hi:
		return Lo
	}
	return A
}
