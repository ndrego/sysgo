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

func (X LogicValue) Unary(op rune) LogicValue {
	switch op {
	case '~':
		switch X {
		case Lo:
			return Hi
		case Hi:
			return Lo
		default:
			return X
		}
	default:
		return X
	}
}
