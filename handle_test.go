package funker

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleFunction(t *testing.T) {
	type addArgs struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	done := make(chan bool, 1)
	errors := make(chan error, 1)

	go func() {
		err := Handle(func(args *addArgs) int {
			return args.X + args.Y
		})
		if err != nil {
			errors <- err
		}
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		ret, err := Call("localhost", addArgs{X: 1, Y: 2})
		if err != nil {
			errors <- err
			return
		}
		assert.Equal(t, ret, 3.0)
		close(done)
	}()

	select {
	case <-done:
	case err := <-errors:
		t.Fatal(err)
	}
}

func TestSecondRequest(t *testing.T) {
	type data struct {
		X string `json:"x"`
	}
	handlerTakes := 3 * time.Second
	epsilon := 100 * time.Millisecond
	go func() {
		err := Handle(func(args *data) *data {
			t.Logf("Handling %+v (takes %s)", args, handlerTakes)
			time.Sleep(handlerTakes)
			return args
		})
		if err != nil {
			t.Fatal(err)
		}
	}()

	call1Done := make(chan interface{}, 1)
	t.Logf("1st call begins in background, this should pass")
	go func() {
		res, err := Call("localhost", data{X: "shouldPass"})
		if err != nil {
			t.Fatal(err)

		}
		call1Done <- res
	}()
	t.Logf("sleeping for %s, so that make sure the 1st call request is received", epsilon)
	time.Sleep(epsilon)

	t.Logf("2nd call begins in foreground, this should fail immediately")
	call2Begin := time.Now()
	res, err := Call("localhost", data{X: "call2"})
	if err == nil {
		t.Fatalf("2nd call should fail, got %v", res)
	}
	call2Took := time.Now().Sub(call2Begin)
	t.Logf("2nd call took %s, got err %v", call2Took, err)
	if call2Took > epsilon {
		t.Fatalf("2nd call should have failed immediately (%s), took %s",
			epsilon, call2Took)
	}

	res = <-call1Done
	t.Logf("1st call done, got %+v", res)
}
