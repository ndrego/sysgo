package sysgo

import (

)

type DriverInterface interface {
	GetValue() LogicValue
	GetLastValue() LogicValue
}

type ReceiverInterface interface {
	SetValue(LogicValue)
}
