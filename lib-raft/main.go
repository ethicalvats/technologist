package main

import (
	"fmt"
	"lib-raft/labrpc"
	"lib-raft/raft"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

type RaftEngine struct {
	raft      *raft.Raft
	persistor *raft.Persister
}

func (e *RaftEngine) run(i int) []*labrpc.Node {
	e.persistor = raft.MakePersister()
	applyCh := make(chan raft.ApplyMsg)

	count := 0

	go func(c int) {
		for m := range applyCh {
			fmt.Printf("%d ch msg %v\n", i, m)
			c++
			fmt.Println(strconv.Itoa(c) + " total items")
		}
	}(count)

	nodes := make([]*labrpc.Node, 3)

	nodes[0] = &labrpc.Node{}
	nodes[0].Endpoint = "9001"
	nodes[0].Id = 0

	nodes[1] = &labrpc.Node{}
	nodes[1].Endpoint = "9002"
	nodes[1].Id = 1

	nodes[2] = &labrpc.Node{}
	nodes[2].Endpoint = "9003"
	nodes[2].Id = 2

	rf := raft.Make(nodes, nodes[i], e.persistor, applyCh)
	e.raft = rf

	return nodes
}

func main() {

	fmt.Println(os.Args)

	server := os.Args[1]
	n, _ := strconv.Atoi(server)

	e := &RaftEngine{}
	nodes := e.run(n)

	rpc.Register(e.raft)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "localhost:"+os.Args[2])
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go http.Serve(l, nil)

	gateway := &raft.Gateway{}
	gateway.Nodes = nodes
	gateway.Connected = false

	gateway.HandleAndRedirect(nodes, ":9000")
}
