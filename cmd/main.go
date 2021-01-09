package cmd

import (
	"joycon"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

func calc(v float32) int {
	r := 500 - int(500*v)
	if r < 0 {
		r = 0
	}
	if r > 999 {
		r = 999
	}
	return r
}

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	devices, err := joycon.Search()
	if err != nil {
		log.Fatalln(err)
	}
	jcs := []*joycon.Joycon{}
	for _, dev := range devices {
		jc, err := joycon.NewJoycon(dev.Path, false)
		if err != nil {
			log.Fatalln(err)
		}
		jcs = append(jcs, jc)
	}
	defer func() {
		for _, jc := range jcs {
			jc.Close()
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	done := make(chan struct{})
	go func() {
		<-sig
		close(done)
	}()


	var wg sync.WaitGroup
	for _, jc := range jcs {
		wg.Add(1)
		go func(jc *joycon.Joycon) {
			defer wg.Done()
			states := []joycon.State{}
			sensors := []joycon.Sensor{}
			tick := time.NewTicker(50 * time.Millisecond)
			for {
				select {
				case <-tick.C:
					log.Printf("%s states:%d, sensors:%d", jc.Name(), len(states), len(sensors))

					continue
				case <-done:
					return
				case s, ok := <-jc.State():
					if !ok {
						return
					}
					states = append(states, s)

						log.Printf("state: %s %3d:%3d%% %06X %v%v",
							jc.Name(),
							s.Tick, s.Battery, s.Buttons, s.LeftAdj, s.RightAdj,
						)

				case s, ok := <-jc.Sensor():
					if !ok {
						return
					}
					sensors = append(sensors, s)

						log.Printf("sensor: %s %3d:%v%v",
							jc.Name(),
							s.Tick, s.Accel, s.Gyro,
						)

				}
			}
		}(jc)
	}
	wg.Wait()
}
