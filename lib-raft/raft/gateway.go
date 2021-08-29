package raft

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

		buf, bodyErr := ioutil.ReadAll(r.Body)
		if bodyErr != nil {
			fmt.Println("bodyErr ", bodyErr.Error())
			http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
			return
		}

		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		fmt.Printf("BODY: %q", rdr1)
		r.Body = rdr2

		decoder := json.NewDecoder(r.Body)
		todo := &Todo{}
		err := decoder.Decode(&todo)
		if err != nil {
			fmt.Println("http decoder error ", err)
		}
		todobytes, _ := json.Marshal(todo)
		args := &StartArgs{Data: todobytes}
		for _, s := range servers {
			reply := &StartReply{}
			gob.Register(StartArgs{})
			ok := s.Call("Raft.Start", args, reply)
			// fmt.Println(ok)
			if ok && reply.IsLeader {
				b := &bytes.Buffer{}
				for _, l := range reply.Log {
					// val, _ := json.Marshal(l.Command)
					// res = res + string(val)
					encoder := json.NewEncoder(b)
					err := encoder.Encode(l.Command)
					if err != nil {
						fmt.Println("json encoding error ", err)
					}
				}
				io.WriteString(w, b.String())
			}
		}
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
