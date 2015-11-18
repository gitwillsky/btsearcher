package dht

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

var BOOTSTRAP []string = []string{
	"router.magnets.im:6881", // dht.transmissionbt.com
	"router.bittorrent.com:6881", //dht.transmissionbt.com
	"dht.transmissionbt.com:6881", // router.bittorrent.com
	//"82.221.103.244:6881", // router.utorrent.com
}

type ID []byte

// Generate Node ID
func GenerateNodeId() ID {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	h := sha1.New()

	io.WriteString(h, time.Now().String())
	io.WriteString(h, string(random.Int()))
	return h.Sum(nil)
}

// Convert ID to string
func (i ID) String() string {
	return hex.EncodeToString(i)
}

// ID sum
func (id ID) Sum() int {
	var d int

	for i := 0; i < len(id); i++ {
		d += int(id[i])
	}

	return d
}

// node contact info.
type ContactInfo struct {
	IP         net.IP    // ip address
	Port       int       // UDP port
	Id         ID        // node ID
	lastAccess time.Time // last access time
}

// DHT node
type Node struct {
	info    *ContactInfo
	routing *Routing // node routing table
	network *Network
	krpc    *KRPC
	output  chan string
}

// network
type Network struct {
	addr *net.UDPAddr
	conn *net.UDPConn
}

func NewDHTNode(outputChan chan string, port int) *Node {
	var err error

	node := &Node{
		info: &ContactInfo{
			Id: GenerateNodeId(),
		},
	}

	// new routing table
	node.routing = NewRouting(node)

	// init udp network connection.
	node.network = new(Network)
	udpAddr := new(net.UDPAddr)
	udpAddr.Port = port
	if node.network.conn, err = net.ListenUDP("udp", udpAddr); err != nil {
		panic(err.Error())
	}

	laddr := node.network.conn.LocalAddr().(*net.UDPAddr)
	node.info.IP = laddr.IP
	node.info.Port = laddr.Port
	// new KRPC
	node.krpc = NewKRPC()
	// output channel
	node.output = outputChan

	return node
}

// create routing table
func (n *Node) CreateTable() error {
	for {
		if n.routing.buckets[0].Len() == 0 {
			for _, host := range BOOTSTRAP {
				addr, err := net.ResolveUDPAddr("udp", host)
				if err != nil {
					return err
				}

				queryNode := &ContactInfo{
					IP:   addr.IP,
					Port: addr.Port,
				}
				n.Find_Node(queryNode, n.info.Id)
			}
		} else {
			for _, node := range n.routing.buckets[0].Nodes {

				t := time.Now()
				d, _ := time.ParseDuration("-10s")
				last := t.Add(d)

				if node.lastAccess.Before(last) {
					continue
				}

				n.Find_Node(node, GenerateNodeId())
			}

			n.routing.buckets[0].Nodes = nil

			time.Sleep(2 * time.Second)
		}
	}
}

// find_node
func (n *Node) Find_Node(queryingNodeInfo *ContactInfo, target ID) error {
	if queryingNodeInfo.IP.Equal(net.IPv4(0, 0, 0, 0)) || queryingNodeInfo.Port == 0 {
		return errors.New("Can not parse Bootstrap node address.")
	}

	addr := new(net.UDPAddr)
	addr.IP = queryingNodeInfo.IP
	addr.Port = queryingNodeInfo.Port

	args := make(map[string]interface{})
	args["target"] = string(target)
	args["id"] = string(n.info.Id)
	dat, err := n.krpc.EncodeQueryPackage("find_node", args)
	if err != nil {
		return err
	}
	if _, err := n.network.conn.WriteToUDP(dat, addr); err != nil {
		return err
	}

	return nil
}

// run
func (n *Node) Run() {
	fmt.Printf("My own node id is %s \n", n.info.Id)
	fmt.Printf("BTSearcher is running at %s \n", n.network.conn.LocalAddr().String())

	go func() {
		err := n.CreateTable()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	go func() {
		n.GetInformation()
	}()
}

// get response/request from other node
func (n *Node) GetInformation() {
	b := make([]byte, 1000)

	for {
		_, addr, err := n.network.conn.ReadFromUDP(b)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}

		msg, err := n.krpc.DecodePackage(b)
		if err != nil {
			//fmt.Println("Error: " + err.Error())
			continue
		}

		// query message
		if query, ok := msg.(*QueryMsg); ok {
			// query node.
			queryNode := &ContactInfo{
				IP:   addr.IP,
				Port: addr.Port,
				Id:   ID(query.A["id"].(string)),
			}

			// response
			response := make(map[string]string)

			switch query.Q {
			case "ping":
				//fmt.Printf("Receive a ping. \n")

				response["id"] = string(n.info.Id)
				dat, err := n.krpc.EncodeResponsePackage(query.tid, response)
				if err != nil {
					//fmt.Println("Error: " + err.Error())
					continue
				}
				if _, err := n.network.conn.WriteToUDP(dat, addr); err != nil {
					fmt.Println("Error: " + err.Error())
					continue
				}
			case "find_node":
				//fmt.Printf("Receive a find_node. \n")
				closeNodes := n.routing.buckets[1].Nodes
				nodes := convertByteStream(closeNodes)
				response["id"] = string(n.info.Id)
				response["nodes"] = bytes.NewBuffer(nodes).String()
				dat, err := n.krpc.EncodeResponsePackage(query.tid, response)
				if err != nil {
					//fmt.Println("Error: " + err.Error())
					continue
				}
				if _, err := n.network.conn.WriteToUDP(dat, addr); err != nil {
					fmt.Println("Error: " + err.Error())
					continue
				}
			case "announce_peer":
				//fmt.Printf("Receive a announce_peer. ")
				if infohash, ok := query.A["info_hash"].(string); ok {
					n.output <- ID(infohash).String() + ":" + addr.IP.String()
				}
			case "get_peers":
				//fmt.Printf("Receive a get_peers. ")
				if _, ok := query.A["info_hash"].(string); ok {
					//n.output <- ID(infohash).String()

					nodes := convertByteStream(n.routing.buckets[1].Nodes)
					response["id"] = string(n.info.Id)
					response["token"] = n.GenerateToken(queryNode)
					response["nodes"] = bytes.NewBuffer(nodes).String()
					dat, err := n.krpc.EncodeResponsePackage(query.tid, response)
					if err != nil {
						//fmt.Println("Error: " + err.Error())
						continue
					}
					if _, err := n.network.conn.WriteToUDP(dat, addr); err != nil {
						fmt.Println("Error: " + err.Error())
						continue
					}
				}
			}
			n.routing.InsertNode(queryNode)
		} else if response, ok := msg.(*ResponseMsg); ok {
			// 判断tid,如果不是我们请求的包,直接抛弃.
			/*tid, err := strconv.Atoi(response.tid)
			if err != nil || !n.krpc.tid.Have(uint32(tid)) {
				fmt.Println("Received a wrong transaction id message.")
				continue
			}
			*/
			// response message.
			if nodesStr, ok := response.R["nodes"].(string); ok {
				nodes := parseBytesStream([]byte(nodesStr))
				for _, v := range nodes {
					//fmt.Printf("Found new node, id = %s   ip = %s   port = %d \n", v.Id,v.IP,v.Port)
					n.routing.InsertNode(v)
				}
			}
		} else if _, ok := msg.(*ErrorMsg); ok {
			//fmt.Printf("Received a error message. errors = %v \n", err.Errors)
		}
	}
}

// Generate token
func (n *Node) GenerateToken(sender *ContactInfo) string {
	h := sha1.New()
	io.WriteString(h, sender.IP.String())
	io.WriteString(h, time.Now().String())

	token := bytes.NewBuffer(h.Sum(nil)).String()

	return token
}

func convertByteStream(nodes []*ContactInfo) []byte {
	buffer := bytes.NewBuffer(nil)

	for _, v := range nodes {
		buffer.Write(v.Id)
		buffer.Write(v.IP.To4())
		buffer.WriteByte(byte((v.Port & 0xFF00) >> 8))
		buffer.WriteByte(byte(v.Port & 0xFF))
	}

	return buffer.Bytes()
}

func parseBytesStream(data []byte) []*ContactInfo {
	var nodes []*ContactInfo

	for j := 0; j < len(data); j = j + 26 {
		if j + 26 > len(data) {
			break
		}

		kn := data[j : j + 26]
		node := &ContactInfo{
			Id:   ID(kn[0:20]),
			IP:   kn[20:24],
			Port: int(kn[24:26][0]) << 8 + int(kn[24:26][1]),
		}

		nodes = append(nodes, node)
	}

	return nodes
}
