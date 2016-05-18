package sysgo

import (
)

//go:generate stringer -type LogicState

type LogicState uint8

const (
	Lo LogicState = iota
	Hi
	HiZ
	Undefined
)


func CombineLogicState(cur, next LogicState) LogicState {
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

func (X LogicState) Unary(op rune) LogicState {
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
