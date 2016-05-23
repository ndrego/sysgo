package sysgo

import (
	"fmt"
)

type Register struct {
	Name string
	currentValue *Value1
	nextValue *Value1
	lastValue *Value1
	modified bool
}

func (A *Register) updateValue() {
	if A.modified {
		A.lastValue = A.currentValue
		A.currentValue = A.nextValue
		A.modified = false
	}

}

func (A *Register) SetValue(v ValueInterface) error {
	if A.modified {
		return fmt.Errorf("Setting register %s multiple times in same event.", A.Name)
	}

	switch v := v.(type) {
	case *Value1:
		A.nextValue = v
	case *Value64, *ValueBig:
		n := NewValue(1).(*Value1)
		n.SetBit(0, v.GetBit(0))
		A.nextValue = n
	}
	A.modified = true

	return nil
}

func (A *Register) GetValue() ValueInterface {
	return A.currentValue
}

func (A *Register) GetLastValue() ValueInterface {
	return A.lastValue
}

func NewRegister(name string) (r *Register) {
	r = new(Register)
	r.Name = name
	r.currentValue = NewValue(1).(*Value1)
	r.nextValue = NewValue(1).(*Value1)
	r.lastValue = NewValue(1).(*Value1)

	r.currentValue.SetBit(0, Undefined)
	r.nextValue.SetBit(0, Undefined)
	r.lastValue.SetBit(0, Undefined)

	r.modified = false

	return
}
