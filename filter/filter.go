package filter
import (
	"io/ioutil"
	"regexp"
)

var keywordReg *regexp.Regexp

// init keyword from conf/key.txt
func init() {
	content, err := ioutil.ReadFile("conf/key.txt")
	if err != nil {
		panic("Read keywords error " + err.Error())
	}

	keywordReg, err = regexp.Compile(string(content))
	if err != nil {
		panic("Compile filter keywords error " + err.Error())
	}
}

//  非法关键字过滤
func IsIllegalWords(name string) bool {
	return keywordReg.MatchString(name)
}