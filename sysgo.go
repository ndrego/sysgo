package sysgo

import (

)

type DriverInterface interface {
	GetValue() Value
	GetLastValue() Value
}

type ReceiverInterface interface {
	SetValue(Value)
}
