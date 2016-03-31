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

	init := func(c chan<- sysgo.ProceduralBlockEvent) {
		fmt.Printf("Setting clk initial value to 0\n")
		clk.SetValue(sysgo.Lo)

		c <- sysgo.Complete
	}

	ssFunc := func() error {
		// sysgo.Delay(5)
		clk.SetValue(clk.GetValue().Invert())

		return nil
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
	sim.Initialize(1e-9, 1e-12)
	sim.RegisterModule(&tm.Module)
	sim.Run()
}
