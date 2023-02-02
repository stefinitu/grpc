package main

import (
	"myproj/core"
	"os"
)

func main() {
	nodeType := os.Args[1]
	switch nodeType {
	case "master":
		core.GetMasterNode().Start()
	case "worker":
		core.GetWorkerNode().Start()
	case "masterworker":
		if os.Args[2] == "1" || os.Args[2] == "2" {
			node2Type := os.Args[2]
			core.GetMasterWorkerNode(node2Type).Start(node2Type)
		}
	default:
		panic("invalid node type")
	}

}
