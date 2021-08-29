package labrpc

import (
	"log"
	"net/rpc"
)

type Node struct {
	Endpoint string
	Id       int
}

func (n *Node) Call(svcMeth string, args interface{}, reply interface{}) bool {

	client, err := rpc.DialHTTP("tcp", "localhost:"+n.Endpoint)
	if err != nil {
		// log.Println("client err: ", err)
		return false
	}
	defer client.Close()
	if err != nil {
		// log.Println("dialing error:", err)
		return false
	}
	err = client.Call(svcMeth, args, reply)
	if err != nil {
		log.Println("raft error:", err)
		return false
	}

	return true
}
