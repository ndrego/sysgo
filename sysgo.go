package sysgo

import (

)

type DriverInterface interface {
	GetValue() Value
}

type ReceiverInterface interface {
	SetValue(Value)
}

type Port struct {

}
