package dht
import "testing"

// test id sum
func Test_IdSum(t *testing.T) {
	s := GenerateNodeId()

	t.Log(s.Sum())
}
