package filter
import "testing"


// test filter illegal words
func Test_FilterIllegalWords(t *testing.T) {
	name := "反党党ssss"

	if IsIllegalWords(name) == false {
		t.Error("Test filter illegal words failed!")
	}
}


