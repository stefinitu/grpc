package main

import (
	"myproj/core"
	"os"
	"strconv"
)

func main() {
	nodeType := os.Args[1]
	switch nodeType {
	case "master":
		portValue := os.Args[2]
		core.GetMasterNode(portValue).Start(portValue)
	case "worker":
		portValue := os.Args[2]
		core.GetWorkerNode(portValue).Start(portValue)
	case "masterworker":
		if os.Args[2] == "1" || os.Args[2] == "2" {
			node2Type := os.Args[2]
			port := os.Args[3]
			nrDials := os.Args[4]
			portDialArray := []string{}
			nrDialsInt, _ := strconv.Atoi(nrDials)
			for i := 0; i < nrDialsInt; i++ {
				portDialArray = append(portDialArray, os.Args[5+i])
			}
			core.GetMasterWorkerNode(node2Type, port, nrDials, portDialArray).Start(node2Type, port, nrDials, portDialArray)
		}
	default:
		panic("invalid node type")
	}
}
