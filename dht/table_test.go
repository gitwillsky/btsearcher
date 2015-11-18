package dht
import "testing"


func Test_GetDistance(t *testing.T) {
	a := &ContactInfo{
		Id:GenerateNodeId(),
	}

	b := &ContactInfo{
		Id:GenerateNodeId(),
	}

	t.Log(getDistance(a, b))
}