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

// This is essentially a wire-OR
func (A LogicState) Combine(l LogicState) LogicState {
	if A == l {
		return A
	} else if A == Undefined || l == Undefined {
		return Undefined
	} else if A == HiZ && (l == Lo || l == Hi) {
		return l
	} else if (A == Lo || A == Hi) && l == HiZ {
		return A
	}

	return Undefined
}

func (A LogicState) Rune() rune {
	switch A {
	case Lo:
		return '0'
	case Hi:
		return '1'
	case HiZ:
		return 'z'
	case Undefined:
		return 'x'
	}
	return 'x'
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
