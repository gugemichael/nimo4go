package nimo

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

func GoVarLoop(n uint64, function func()) {
	for n != 0 {
		function()
		n--
	}
}
