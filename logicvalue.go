package sysgo

import (
)

//go:generate stringer -type LogicValue

type LogicValue uint8

const (
	Lo LogicValue = iota
	Hi
	HiZ
	Undefined
)


func CombineLogicValue(cur, next LogicValue) LogicValue {
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

func (A LogicValue) UnaryOp(op rune) LogicValue {
	switch op {
	case '~':
		return A.Invert()
	default:
		return A
	}
}

func (A LogicValue) Invert() LogicValue {
	switch {
	case A == Lo:
		return Hi
	case A == Hi:
		return Lo
	}
	return A
}
