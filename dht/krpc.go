package dht

import (
	"github.com/zeebo/bencode"
	"errors"
	"sync/atomic"
	"math"
	"fmt"
)

// KRPC message
type Msg struct {
	T string // transaction id
	Y string // the type of the message
}

// KRPC queries message
type QueryMsg struct {
	tid string
	Q   string                 // method name of the query.
	A   map[string]interface{} // method arguments.
}

// KRPC response message
type ResponseMsg struct {
	tid string
	R   map[string]interface{} // return values.
}

// KRPC Error message
type ErrorMsg struct {
	tid    string
	Errors []interface{}
}

// KRPC Protocol
type KRPC struct {
	tid uint32
}

// New KRPC
func NewKRPC() *KRPC {
	result := &KRPC{}
	//go result.tid.GC()

	return result
}

/*
type tidStorage struct {
	id        uint32
	container map[uint32]string
	lock      sync.Mutex
}

// tid gc
func (ts *tidStorage) GC() {
	for {
		now := time.Now()
		d, _ := time.ParseDuration("-10s")
		t := now.Add(d)

		for k, v := range ts.container {
			kTime, _ := time.Parse(time.RFC3339, v)
			if kTime.Before(t) {
				ts.lock.Lock()
				delete(ts.container, k)
				ts.lock.Unlock()
			}
		}
	}
}

// is transaction id in tid container?
func (ts *tidStorage) Have(tid uint32) bool {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	if _, ok := ts.container[tid]; ok {
		return true
	}
	return false
}

*/
// Generate transaction id
func (k *KRPC) GenerateTID() uint32 {
	k.tid = atomic.AddUint32(&k.tid, 1) % math.MaxUint16

	return k.tid
}

// Decode KRPC Package
func (k *KRPC) DecodePackage(b []byte) (interface{}, error) {
	dat := make(map[string]interface{})

	if err := bencode.DecodeBytes(b, &dat); err != nil {
		return nil, err
	}

	message := &Msg{}
	var ok bool
	if message.T, ok = dat["t"].(string); !ok {
		return nil, errors.New("Response package not have transcation ID.")
	}

	if message.Y, ok = dat["y"].(string); !ok {
		return nil, errors.New("Response package message type unknown.")
	}

	// encode message type.
	switch message.Y {
	// query
	case "q":
		query := &QueryMsg{}
		query.tid = message.T
		query.Q = dat["q"].(string)
		query.A = dat["a"].(map[string]interface{})
		return query, nil
	// response
	case "r":
		response := &ResponseMsg{}
		response.tid = message.T
		response.R = dat["r"].(map[string]interface{})
		return response, nil
	// error
	case "e":
		err := &ErrorMsg{}
		err.tid = message.T
		err.Errors = dat["e"].([]interface{})
		return err, nil
	default:
		return nil, errors.New("Can not parse message type.")
	}
}

// Encode KRPC response package
func (k *KRPC) EncodeResponsePackage(tid string, response map[string]string) ([]byte, error) {
	dat := make(map[string]interface{})
	dat["t"] = tid
	dat["y"] = "r"
	dat["r"] = response

	return bencode.EncodeBytes(dat)
}

// Encode KRPC Query package
func (k *KRPC) EncodeQueryPackage(methodName string, args map[string]interface{}) ([]byte, error) {
	dat := make(map[string]interface{})
	dat["t"] = fmt.Sprintf("%d", k.GenerateTID())
	dat["y"] = "q"
	dat["q"] = methodName
	dat["a"] = args

	return bencode.EncodeBytes(dat)
}