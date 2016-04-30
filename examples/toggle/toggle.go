package main

import (
	"fmt"
	"github.com/ndrego/sysgo"
)

type ToggleModule struct {
	sysgo.Module
}

func NewToggleModule() (tm *ToggleModule) {
	clk := sysgo.NewRegister("clk")

	init := func(cp *sysgo.SimChanPair) (bool, error){
		fmt.Printf("Setting clk initial value to 0\n")
		clk.SetValue(sysgo.Lo)
		sysgo.Delay(cp, 100)

		return true, nil // Indicates the sim should finish
	}

	ssFunc := func(cp *sysgo.SimChanPair) (bool, error) {
		waitComplete := sysgo.Delay(cp, 5)
		if !waitComplete {
			return true, nil
		}

		fmt.Printf("clk toggle @ %d: %d\n", sysgo.SimTime(), clk.GetValue())
		clk.SetValue(clk.GetValue().Invert())

		return false, nil
	}

	sc := sysgo.NewSensitivityClause(ssFunc)
	
	tm = new(ToggleModule)
	tm.Name = "Toggler"
	tm.Registers = []*sysgo.Register{clk}
	tm.Initializers = []sysgo.InitializerFunc{init}
	tm.SensitivityClauses = []*sysgo.SensitivityClause{sc}

	return 
}

func main() {
	tm := NewToggleModule()
	sim := sysgo.GetSimulator()
	sim.Initialize(1e-9, 1e-9)
	sim.RegisterModule(&tm.Module)
	sim.Run()
}
