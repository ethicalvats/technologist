package raft

import (
	"encoding/json"
	"fmt"
	"lib-raft/labrpc"
	"net/http"
	"time"
)

type Todo struct {
	Text string `json:"text"`
	Id   string `json:"id"`
}

type Gateway struct {
	Nodes     []*labrpc.Node
	Connected bool
}

func (g *Gateway) HandleAndRedirect(servers []*labrpc.Node, port string) {

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		todo := &Todo{}
		err := decoder.Decode(&todo)
		if err != nil {
			fmt.Println("http decoder error ", err)
		}
		bytes, _ := json.Marshal(todo)
		args := &StartArgs{Data: bytes}
		for _, s := range servers {
			reply := &StartReply{}
			ok := s.Call("Raft.Start", args, &reply)
			fmt.Println(reply)
			fmt.Println(ok)
		}

		fmt.Println(todo)
	})

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		err := http.ListenAndServe(port, nil)
		if err != nil {
			// fmt.Println("http listen error", err)
			g.Connected = false
		}
		g.Connected = true
	}
}
