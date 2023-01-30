package main

import (
	"core"
	"fmt"
	"os"
)

func main() {
	nodeType := os.Args[1]
	fmt.Println(nodeType)
	switch nodeType {
	case "master":
		core.GetMasterNode().Start()
	case "worker":
		core.GetWorkerNode().Start()
	default:
		panic("invalid node type")
	}
}
