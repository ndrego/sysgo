package sysgo

import (
)

type Port struct {
	Name string
	driver DriverInterface
	receiver ReceiverInterface
}

func (A *Port) GetValue() LogicValue {
	return A.driver.GetValue()
}

func (A *Port) GetLastValue() LogicValue {
	return A.driver.GetLastValue()
}
