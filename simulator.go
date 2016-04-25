package sysgo

import (
	"fmt"
	_ "math"
	"sync"
)

type ProceduralBlockEvent uint8
const (
	Complete ProceduralBlockEvent = iota
	BlockOnDelay
	SimFinish
)

type simInternalEvent uint32
const (
	allInitsComplete simInternalEvent = iota
	zeroTimeInitsComplete
	proceduralBlockError
	simFinish
)

type SimulatorEventType uint32
const (
	Timestep SimulatorEventType = iota
	SensitivityMatch
)

type SimulatorEvent struct {
	Type SimulatorEventType
	Data interface{}
}

type Simulator struct {
	timescale float64
	precision float64
	modules []*Module

	simTime uint64
	
	// Used to tell if an initializer block is complete or just
	// waiting on a delay.
	initBlock chan ProceduralBlockEvent

	// Internal simulator event channel
	simChan chan simInternalEvent

	delayChannels []chan SimulatorEvent
	// senseChannelMap map[string]chan Event

	numInitBlocks uint32
	numSenseClauses uint32
}

// Make sure there is only a single simulator instance using sync.Once
var simulator *Simulator
var once sync.Once

func GetSimulator() *Simulator {
    once.Do(func() {
        simulator = &Simulator{}
    })
    return simulator
}

func (A *Simulator) Initialize(timescale, precision float64) {
	A.timescale = timescale
	A.precision = precision
	A.modules = make([]*Module, 0, 10)
	A.delayChannels = make([]chan SimulatorEvent, 0, 10)
}

// This should only be called once per *TOP-LEVEL* module
func (A *Simulator) RegisterModule(m *Module) {
	A.modules = append(A.modules, m)

	// Figure out the total # of initializer and procedural blocks
	n, s := m.getNumBlocks()
	A.numInitBlocks += n
	A.numSenseClauses += s
}

func (A *Simulator) Run() {
	A.simTime = 0
	/* Order of operations:
           1. Go through all initialize clauses and fire them off, each into it's own go routine
           2. Setup go routines for all sensitivity clauses with channels to each for event broadcast.
           3. (Except on iteration 0) Update all registers currentValue <- nextValue, keeping track of what's changed
           4. Update all wires, by propagating driver values to receivers, keeping track of what's changed.
           5. Determine active clauses based on registers/wires that have changed and send events to each go routine.

        */

	A.simChan = make(chan simInternalEvent, 10) // Should this be changed to some other value?
	go A.runInitializers()
	// Need to wait
	finished := A.waitOnInitializers()

	if finished {
		fmt.Printf("Simulation complete at time %0.6f\n", A.simTime)
		return
	}

	// Update registers, wires, ports.
	// May want to parallelize this operation in the future, but
	// need to be careful to organize it correctly.
	wg := new(sync.WaitGroup)
	wg.Add(len(A.modules))
	for _, m := range A.modules {
		go A.updateRegisters(m, wg)
	}
	wg.Wait()

	wg.Add(len(A.modules))
	for _, m := range A.modules {
		go A.propagateWires(m, wg)
	}
	wg.Wait()
}

func (A *Simulator) waitOnInitializers() bool {

WaitLoop:
	for {
		event, ok := <- A.simChan
		if !ok {
			close(A.simChan)
			return false
		} else {
			switch event {
			case allInitsComplete, zeroTimeInitsComplete:
				// Can proceed now
				break WaitLoop
			case simFinish:
				return true
			}
		}
	}

	return false
}


func (A *Simulator) runInitializers() {
	A.initBlock = make(chan ProceduralBlockEvent, A.numInitBlocks)

	fmt.Printf("There are %d registered modules\n", len(A.modules))
	for _, m := range A.modules {
		iFuncs := m.getAllInitializers()
		for _, i := range iFuncs {
			go i(A.initBlock)
		}
	}

	// Process the init block events
	initsComplete := uint32(0)
	blocksOnDelays := uint32(0)
Loop:
	for {
		event, ok := <-A.initBlock
		if !ok {
			close(A.simChan)
			return
		} else {
			switch event {
			case Complete:
				initsComplete += 1
			case BlockOnDelay:
				blocksOnDelays += 1
			case SimFinish:
				A.simChan <- simFinish
				break Loop
			}

			if initsComplete == A.numInitBlocks {
				A.simChan <- allInitsComplete
				break Loop
			}
			if initsComplete + blocksOnDelays == A.numInitBlocks {
				A.simChan <- zeroTimeInitsComplete
			}
		}
		
	}
	fmt.Printf("initsComplete: %d\n", initsComplete)
}

/* func Delay(d float64, c chan<- ProceduralBlockEvent) {
	sim := GetSimulator()

	// Round d to precision
	var del uint64
	del = uint64(math.Floor(d / sim.precision + 0.5))

	targetTime = sim.simTime + del
} */

func (A *Simulator) updateRegisters(m *Module, wg *sync.WaitGroup) {
	subWG := new(sync.WaitGroup)
	subWG.Add(len(m.SubModules) + 1)
	for _, sm := range m.SubModules {
		go A.updateRegisters(sm, subWG)
	}

	// First go through the module's registers
	for _, r := range m.Registers {
		if r.modified {
			// TODO: Need to let any sensitivity clauses know
			// that there is a change. And record the change
			// if we're recording this module.
			r.lastValue = r.currentValue
			r.currentValue = r.nextValue
			fmt.Printf("Set %s.%s (%d -> %d). Simtime: %d\n", m.Name, r.Name, r.lastValue, r.currentValue, A.simTime)
			r.modified = false
		}
	}
	subWG.Done()
	subWG.Wait()
	
	wg.Done()
}

func (A *Simulator) propagateWires(m *Module, wg *sync.WaitGroup) {
	subWG := new(sync.WaitGroup)
	subWG.Add(len(m.SubModules) + 1)
	for _, sm := range m.SubModules {
		go A.propagateWires(sm, subWG)
	}

	for _, w := range m.Wires {
		w.computeValue()
	}
	subWG.Done()
	subWG.Wait()
	
	wg.Done()
}
