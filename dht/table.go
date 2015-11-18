package dht
import (
	"errors"
)


//const BucketMaxItem = 100

// dht routing
type Routing struct {
	selfNode *Node
	buckets  []*Bucket
}

// new routing
func NewRouting(node *Node) *Routing {
	routing := &Routing{}

	routing.selfNode = node
	routing.buckets = make([]*Bucket, 2)
	routing.buckets[0] = NewBucket()
	routing.buckets[1] = NewBucket()

	return routing
}

/*
// kad distance.
func getDistance(a *ContactInfo, b *ContactInfo) int {
	var distance int

	// 20 bit id.
	for i := 0; i < 20; i++ {
		distance = distance + int(a.Id[i] ^ b.Id[i])
	}

	return distance
}

// which bucket
func (r *Routing) WhickBucket(node *ContactInfo) *Bucket {
	for _, bucket := range r.buckets {
		if node.Id.Sum() < bucket.max &&
		node.Id.Sum() > bucket.min {
			return bucket
		}
	}

	// not find , create new one.
	bucket := NewBucket(node.Id.Sum(), node.Id.Sum() + BucketMaxItem)
	r.buckets = append(r.buckets, bucket)

	return bucket
}
*/

// Insert Node
func (r *Routing) InsertNode(node *ContactInfo) error {
	// 如果新插入的node的ID长度不为20位,终止.
	if len(node.Id) != 20 {
		return errors.New("Add node into the routing table failed, the node id lenght invalid.")
	}

	// 如果待插入的是自己本身,则忽略.
	if node.Id.String() == r.selfNode.info.Id.String() {
		return nil
	}

	// 现将1号填满
	if r.buckets[1].Len() < 8 {
		r.buckets[1].Add(node)
	} else {

		// 找到将要插入的bucket.
		//bucket := r.WhickBucket(node)

		// 插入
		//bucket.Add(node)
		r.buckets[0].Add(node)
	}

	return nil
}