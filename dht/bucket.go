package dht

import (
	"errors"
	"time"
)

// DHT bucket
type Bucket struct {
	Nodes      []*ContactInfo
	lastAccess time.Time
}

// New bucket
func NewBucket() *Bucket {
	return &Bucket{
		lastAccess: time.Now(),
	}
}

// bucket len
func (b *Bucket) Len() int {
	return len(b.Nodes)
}

// find node in bucket
func (b *Bucket) FindNode(id ID) (*ContactInfo, error) {
	for _, v := range b.Nodes {
		if v.Id.Sum() == id.Sum() {
			return v, nil
		}
	}

	return nil, errors.New("Can not find this node in the bucket.")
}

// update bucket lastchange time.
func (b *Bucket) UpdateTime(n *ContactInfo) {
	b.lastAccess = time.Now()
	n.lastAccess = time.Now()
}

// add node to bucket
func (b *Bucket) Add(n *ContactInfo) {

	if node, err := b.FindNode(n.Id); err != nil {
		// 不存在
		b.Nodes = append(b.Nodes, n)
	} else {
		// 存在,替换掉.
		node.Id = n.Id
		node.IP = n.IP
		node.Port = n.Port
	}
	// 更新该bucket的新鲜程度.
	b.UpdateTime(n)
}
