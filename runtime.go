package nimo

import "time"

func GoRoutine(function func()) {
	go func() {
		function()
	}()
}

func GoRoutineInLoop(function func()) {
	go func() {
		for {
			function()
		}
	}()
}

func GoRoutineInTimer(duration time.Duration, function func()) {
	go func() {
		for range time.NewTicker(duration).C {
			function()
		}
	}()
}

func GoVarLoop(n uint64, function func()) {
	for n != 0 {
		function()
		n--
	}
}
