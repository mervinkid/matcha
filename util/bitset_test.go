package util_test

import (
	"github.com/mervinkid/allspark/logging"
	"github.com/mervinkid/allspark/util"
	"testing"
)

func TestByteSliceBitSet_Set(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			t.Fatal()
		}
	}()

	logging.SetLogLevel(logging.LInfo)
	bs := util.NewByteSliceBitSet()
	if !bs.IsEmpty() {
		t.Fail()
	}
	bs.Set(1)
	if bs.IsEmpty() {
		t.Fail()
	}
	bs.Clear(100)
	if bs.IsEmpty() {
		t.Fail()
	}
	bs.Clear(1)
	if !bs.IsEmpty() {
		t.Fail()
	}
}
