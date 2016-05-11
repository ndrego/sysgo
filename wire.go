package sysgo

import (
)

type Wire struct {
	Name string
	drivers []DriverInterface
	receivers []ReceiverInterface
	currentValue LogicValue
	lastValue LogicValue
}

func (A *Wire) computeValue() {
	var v LogicValue
	if len(A.drivers) > 0 {
		v = A.drivers[0].GetValue()
		for i := 1; i < len(A.drivers); i++ {
			v = CombineLogicValue(v, A.drivers[i].GetValue())
		}
	} else {
		v = Undefined
	}

	A.lastValue = A.currentValue
	A.currentValue = v
}

func (A *Wire) propagate() {
	A.computeValue()
	for _, r := range A.receivers {
		r.SetValue(A.currentValue)
	}
}

func (A *Wire) GetValue() LogicValue {
	return A.currentValue
}

func (A *Wire) GetLastValue() LogicValue {
	return A.lastValue
}
