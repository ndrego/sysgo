package sysgo

import (
	"fmt"
	"math"
	_ "math/rand"
	"sync"
)

type simInternalEvent uint32
const (
	blockRun simInternalEvent = 1 << iota
	blockComplete
	blockProgress
	blockWait
	blockError
	delayWait	
	updateRegisters
	registerUpdateComplete
	propagateWireValues
	wirePropagateComplete
	simFinish
	allEvents = 0xffffffff
)

type simChanType uint32
const (
	module simChanType = 1 << iota
	initializer
	sensitivity
	allChanTypes = 0xffffffff
)

type eventCounts struct {
	module uint32
	initializer uint32
	sensitivity uint32
}

type SimChanPair struct {
	send chan simInternalEvent // Send from module/initializer/etc. to simulator
	recv chan simInternalEvent // Receive by module/initializer/etc.
	chanType simChanType
	valid bool
	data interface{}
	dataMutex sync.Mutex
}

type Simulator struct {
	timescale float64
	precision float64
	modules []*Module

	simTime uint64
	simTimeMutex sync.Mutex
	
	// Internal simulator event channels
	simChans []*SimChanPair
	simChansMutex sync.Mutex
	simChanCounts map[simChanType]uint

	eventCounts map[simInternalEvent]uint

	runnersWG *sync.WaitGroup
	
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
	A.simChans = make([]*SimChanPair, 0, 10)
	A.simChanCounts = make(map[simChanType]uint)

	A.simChanCounts[module] = uint(0)
	A.simChanCounts[initializer] = uint(0)
	A.simChanCounts[sensitivity] = uint(0)

	A.eventCounts = make(map[simInternalEvent]uint)
}

// This should only be called once per *TOP-LEVEL* module
func (A *Simulator) RegisterModule(m *Module) {
	A.modules = append(A.modules, m)

	// Figure out the total # of initializer and procedural blocks
	n, s := m.getNumBlocks()
	A.numInitBlocks += n
	A.numSenseClauses += s
}

func (A *Simulator) RegisterChannelPair(cp *SimChanPair) (chanId int) {
	chanId = len(A.simChans)
	A.simChansMutex.Lock()
	defer A.simChansMutex.Unlock()

	A.simChans = append(A.simChans, cp)
	A.simChans[chanId].valid = true
	A.simChanCounts[cp.chanType]++
	return chanId
}

func (A *Simulator) UnregisterChannelPair(chanId int) {
	A.simChansMutex.Lock()
	defer A.simChansMutex.Unlock()
	
	A.simChanCounts[A.simChans[chanId].chanType]--
	A.simChans[chanId].valid = false
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

	blockReadyWG := new(sync.WaitGroup)
	A.runnersWG = new(sync.WaitGroup)

	blockReadyWG.Add(int(A.numInitBlocks + A.numSenseClauses))
	A.runnersWG.Add(1)

	go A.spawnRunners(blockReadyWG)

	blockReadyWG.Wait()
		
	// Main event coordination loop
	for finish := false; !finish; {
		A.sendToChannels(initializer | sensitivity, blockRun, false)

		// May need a mutex on simChanCounts
		expEventCount := A.simChanCounts[initializer] + A.simChanCounts[sensitivity]
		// fmt.Printf("simTime: %d, expEventCount: %d\n", A.simTime, expEventCount)

		// Wait for a timestep complete from at least the remaining initializers and
		// all sensitivity clauses
		// fmt.Printf("Waiting for %d channels...\n", expEventCount)		
		A.waitForChannels(allChanTypes, blockProgress | blockWait | blockComplete | delayWait | simFinish, expEventCount)

		finish = A.getEventCounts(simFinish) > 0
		delayCount := A.getEventCounts(delayWait)
		blockWaitCount := A.getEventCounts(blockWait)

		A.sendToChannels(module, updateRegisters, true)
		A.waitForChannels(module, registerUpdateComplete, A.simChanCounts[module])

		A.sendToChannels(module, propagateWireValues, true)
		A.waitForChannels(module, wirePropagateComplete, A.simChanCounts[module])

		// Increment simTime if everything is just waiting
		if (delayCount + blockWaitCount) == expEventCount {
			// Find the minimum target time and fast-forward
			min := uint64(0xffffffffffffffff)
			for _, cp := range A.simChans {
				if cp.valid {
					cp.dataMutex.Lock()
					switch d := cp.data.(type) {
					case nil:
					case uint64:
						tt := d
						if tt > 0 && tt < min {
							min = tt
						}
					}
					cp.dataMutex.Unlock()
				}
			}
			// fmt.Printf("Fast-forwarding to %d\n", min)
			A.simTimeMutex.Lock()
			A.simTime = min
			A.simTimeMutex.Unlock()
		}
	}

	fmt.Printf("Simulator: out of main event loop.\n")

	A.sendToChannels(allChanTypes, simFinish, true)
	
	A.runnersWG.Wait()
}

func (A *Simulator) initEventCounts() {
	for i := blockRun; i <= simFinish; i = i << 1 {
		A.eventCounts[i] = 0
	}
}

func (A *Simulator) getEventCounts(e simInternalEvent) (n uint) {
	n = 0
	if _, ok := A.eventCounts[e]; ok {
		n = A.eventCounts[e]
	}
	return
}

func (A *Simulator) spawnRunners(blockReadyWG *sync.WaitGroup) {
	defer A.runnersWG.Done()
	
	wg := new(sync.WaitGroup)
	wg.Add(len(A.modules))
	
	// Spawn all module runners
	for _, m := range A.modules {
		go m.run(wg, blockReadyWG)
	}

	wg.Wait()
}

func newSimChannelPair(t simChanType) (cp *SimChanPair) {
	cp = new(SimChanPair)
	cp.send = make(chan simInternalEvent, 1)
	cp.recv = make(chan simInternalEvent, 1)
	cp.chanType = t

	return
}

func (A *Simulator) sendToChannels(chanMask simChanType, e simInternalEvent, blocking bool) (c uint) {
	c = 0
	A.simChansMutex.Lock()
	defer A.simChansMutex.Unlock()
	for _, cp := range A.simChans {
		if cp.valid && (cp.chanType & chanMask > 0) {
			// fmt.Printf("Sending %d to 0x%x\n", e, pairId)
			if blocking {
				cp.recv <- e
			} else {
				select {
				case cp.recv <- e:
				default:
				}
			}
			c++
		}
	}
	return
}

func (A *Simulator) waitForChannels(chanMask simChanType, eventMask simInternalEvent, minCount uint) {
	A.initEventCounts()
	n := uint(0)
	for {
		A.simChansMutex.Lock()
		for _, cp := range A.simChans {
			if cp.valid && cp.chanType & chanMask > 0 {
				// Non-blocking read
				select {
				case e := <- cp.send:
					// fmt.Printf("Got event %d on chan 0x%x\n", e, pairId)
					if e & eventMask > 0 {
						A.eventCounts[e]++
						switch e {
						case blockComplete:
							close(cp.recv)
						}
						n++
						//fmt.Printf("Got %d/%d events\n", n, minCount)
					}
				default:
					break
				}
			}
			
		}
		A.simChansMutex.Unlock()
		if n >= minCount {
			break
		}
	}
}

func waitForEvents(c chan simInternalEvent, eventMask simInternalEvent) simInternalEvent {
	for {
		event, ok := <- c
		if !ok {
			close(c)
			return simFinish
		} else {
			if event & eventMask > 0 {
				return event
			}
		}
	}
}

/*
    Waits until the simulator proceeds from the current simTime to 
    simTime + d. Returns true if it waited for the entire delay or false
    if a simFinish event occurred during the delay.
*/
func Delay(cp *SimChanPair, d float64) bool {
	sim := GetSimulator()

	sim.simTimeMutex.Lock()
	simTime := sim.simTime
	sim.simTimeMutex.Unlock()

	// Compute target time by rounding the delay
	
	targetTime := uint64(math.Floor(d * (sim.timescale / sim.precision) + 0.5)) + simTime

	// Catch corner cases
	if targetTime == simTime {
		return true
	}

	// Send a delayWait event
	cp.dataMutex.Lock()
	cp.data = targetTime
	cp.dataMutex.Unlock()
	cp.send <- delayWait

	// Now wait until the correct timestep
	for {
		e := waitForEvents(cp.recv, blockRun | simFinish)
		switch e {
		case simFinish:
			return false
		case blockRun:
			sim.simTimeMutex.Lock()
			if sim.simTime >= targetTime {
				sim.simTimeMutex.Unlock()
				cp.dataMutex.Lock()
				cp.data = uint64(0)
				cp.dataMutex.Unlock()
				return true
			} else {
				select {
				case cp.send <- delayWait:
				default:
				}
			}
			sim.simTimeMutex.Unlock()
		}
	}
}

func SimTime() uint64 {
	sim := GetSimulator()
	sim.simTimeMutex.Lock()
	simTime := sim.simTime
	sim.simTimeMutex.Unlock()
	
	return simTime
}
