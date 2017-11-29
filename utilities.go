package nimo

import "time"


type RateController interface {
	// take one from controller. if no more in the
	// pool (or tokens). make it block.
	//
	// return true. if need to be blocked
	Control() bool
}

// SimpleRateController. simply increase a inner counter
// and block if the counter grew up to it's threshold and
// yield current routine for a while
type SimpleRateController struct {
	Threshold, token int64
	tick             int64
}

func NewSimpleRateController() *SimpleRateController {
	return &SimpleRateController{tick: time.Now().Unix()}
}

func (controller *SimpleRateController) SetThreshold(threshold int64) {
	controller.Threshold = threshold
}

func (controller *SimpleRateController) Control() bool {
	now := time.Now().Unix()
	// current second is forward. we are behind
	if now > controller.tick {
		controller.tick = now
		controller.token = 0
		return false
	}
	AssertTrue(now == controller.tick, "rate controller tick now correct !")

	controller.token++
	return controller.token >= controller.Threshold
}
