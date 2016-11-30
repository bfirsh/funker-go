package funker

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type addArgs struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func TestSimpleFunction(t *testing.T) {
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
