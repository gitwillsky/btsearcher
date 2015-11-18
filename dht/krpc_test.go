package dht
import (
	"testing"
	"time"
)

// test tid storage
func Test_TidStorage(t *testing.T) {
	k := NewKRPC()

	tid := k.tid.GenerateTID()

	time.Sleep(time.Second * 6)

	if k.tid.Have(tid) {
		t.Error("failed")
	}
}
