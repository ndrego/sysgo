package sysgo

import (
)

type Port struct {
	Name string
	driver DriverInterface
	receiver ReceiverInterface
}

func (A *Port) GetValue() Value {
	return A.driver.GetValue()
}

func (A *Port) GetLastValue() Value {
	return A.driver.GetLastValue()
}
