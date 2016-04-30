package sysgo

import (
	"fmt"
)

type Register struct {
	Name string
	currentValue Value
	nextValue Value
	lastValue Value
	modified bool
}

func (A *Register) updateValue() {
	if A.modified {
		A.lastValue = A.currentValue
		A.currentValue = A.nextValue
		A.modified = false
	}

}

func (A *Register) SetValue(v Value) error {
	if A.modified {
		return fmt.Errorf("Setting register %s multiple times in same event.", A.Name)
	}

	A.nextValue = v
	A.modified = true

	return nil
}

func (A *Register) GetValue() Value {
	return A.currentValue
}

func (A *Register) GetLastValue() Value {
	return A.lastValue
}

func NewRegister(name string) (r *Register) {
	r = new(Register)
	r.Name = name
	r.currentValue = Undefined
	r.nextValue = Undefined
	r.lastValue = Undefined
	r.modified = false

	return
}
