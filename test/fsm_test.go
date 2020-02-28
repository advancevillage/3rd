//author: richard
package test

import (
	"fmt"
	"testing"
)

func PayOrder(state string, nextState string) (string, string, error) {

	return state, nextState, nil
}

func CancelOrder(state string, nextState string) (string, string, error) {

	return state, nextState, nil
}

func TestFSM(t *testing.T) {
	fsm := map[string]map[int]func(state string, nextState string) (string, string, error) {
		"ordered:pending_pay": {
			0: PayOrder,
			1: CancelOrder,
		},
	}
	state := "ordered"
	nextState := "pending_pay"
	e := 0
	if _, ok := fsm[fmt.Sprintf("%s:%s", state, nextState)]; !ok {
		t.Log(fmt.Sprintf("don't exist %s and %s", state, nextState))
		return
	}
	if _, ok := fsm[fmt.Sprintf("%s:%s", state, nextState)][e]; !ok {
		t.Log(fmt.Sprintf("don't exist %d event", e))
		return
	}
	handler := fsm[fmt.Sprintf("%s:%s", state, nextState)][e]
	t.Log(handler(state, nextState))
}







