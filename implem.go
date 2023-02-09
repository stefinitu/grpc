package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"time"
)

const (
	marker    string = "marker"
	accept           = "accept"
	reject           = "reject"
	insweep          = "insweep"
	terminate        = "terminate"
)

type Node struct {
	status                   int
	parent                   int
	children                 []int
	unrelated                []int
	self_port                int
	isleaf                   bool
	isparent                 bool
	parent_port              int
	region_color             uint64
	neighbors                []int
	universal_color_set      []int
	terminated_color_set     []int
	neighborhood_color_set   []int
	neighboring_region_ports []int
	incoming_ports           []int
	hasRecordedState         bool
}

func (n Node) RecordLocalState() {
	localStatus := n.status
	n.hasRecordedState = true
	fmt.Println("Local snapshot created with status %d", localStatus)
}

func (n Node) InitiateSnapshot() {
	n.RecordLocalState()
	n.region_color = GenerateNonce()
	SendMarkers(n.region_color, n.neighbors)
	n.parent = n.self_port
	n.universal_color_set = UnionOfSetInteger(int(n.region_color), n.universal_color_set)
	n.InitializeIncomingPortState()
}

func GenerateNonce() uint64 {
	rand.Seed(time.Now().UnixNano())
	var Nonce = rand.Uint64() //Nonce generated (128-bit uint is hard to generate)
	// var stringNonce = fmt.Sprint(Nonce)
	fmt.Println(Nonce)
	return Nonce
}

func (n Node) InitializeIncomingPortState() {
	for _, port := range n.incoming_ports {
		RecordEmptyChannelState(port)
	}
}

func SendMarkers(nonce uint64, ports []int) {
	fmt.Println("Sent %s Message on outgoing ports", marker)
}

func SendInsweep(port int, ncs []int) {
	fmt.Println("Sent %s Message on parent port %d", insweep, port)
}

func SendAccept(port int) {
	fmt.Println("Sent %s on outgoing port %d", accept, port)
}

func SendReject(port int) {
	fmt.Println("Sent %s on outgoing port %d", reject, port)
}

func SendTerminate(nonce uint64, port int, ncs []int) {
	fmt.Println("Sent %s Message to itself on outgoing port %d", terminate, port)
}

func (n Node) RecvMarker(nonce uint64, port int) {
	if n.hasRecordedState == false {
		n.RecordLocalState()
		RecordEmptyChannelState(port)
		SendMarkers(nonce, n.neighbors)
		SendAccept(port)
		n.region_color = nonce
		n.universal_color_set = UnionOfSetInteger(int(n.region_color), n.universal_color_set)
		n.parent = port //parent:=j
	} else {
		RecordChannelState(port)
		SendReject(port)
		if n.region_color != nonce {
			n.neighborhood_color_set = UnionOfSetInteger(int(nonce), n.neighborhood_color_set)
			n.neighboring_region_ports = UnionOfSetInteger(port, n.neighboring_region_ports)
		}
	}
}

func (n Node) AwaitInsweep() {
	//awaits Insweep Messages from each incoming child port
	//for each Insweep Message:
	childport := 0 //mock child port in for loop
	n.neighborhood_color_set = n.RecvInsweep(childport)
}

func (n Node) RecvInsweep(port int) (ncsOutput []int) {
	ncsInputReceived := make([]int, 0) //simularea primirii mesajului Insweep care ar contine parametrul NCS
	n.neighborhood_color_set = UnionOfSets(n.neighborhood_color_set, ncsInputReceived)
	return n.neighborhood_color_set
}

func (n Node) RecvTerminate(port int, nonce uint64, ncs []int) {
	contains := false
	for _, x := range n.terminated_color_set {
		if x == int(nonce) {
			contains = true
		}
	}
	if contains == false {
		n.universal_color_set = UnionOfSets(n.universal_color_set, ncs)
		n.universal_color_set = UnionOfSetInteger(int(nonce), n.universal_color_set)
		n.terminated_color_set = UnionOfSetInteger(int(nonce), n.terminated_color_set)
		sort.Ints(n.universal_color_set)
		sort.Ints(n.terminated_color_set)
		if !reflect.DeepEqual(n.universal_color_set, n.terminated_color_set) { //NOT YET GLOBAL TERMINATION
			portsToSendTerminate := UnionOfSetInteger(n.parent, n.children)
			portsToSendTerminate = UnionOfSets(portsToSendTerminate, n.neighboring_region_ports)
			var portAsArray []int
			portAsArray[0] = port
			portsToSendTerminate = Difference(portsToSendTerminate, portAsArray)

			for _, neighbor_port := range portsToSendTerminate {
				SendTerminate(nonce, neighbor_port, ncs)
			}
		}else{ //GLOBAL TERMINATION
			fmt.Println("GLOBAL TERMINATION")
		}
	}
}

func (n Node) RecvAccept(port int) {
	n.children = UnionOfSetInteger(port, n.children)
	union := UnionOfSets(n.children, n.unrelated)
	var parentAsArray []int
	parentAsArray[0] = n.parent
	difference := Difference(n.neighbors, parentAsArray)
	sort.Ints(union)
	sort.Ints(difference)
	if reflect.DeepEqual(union, difference) {
		// node has received a MARKER on all (incoming) ports
		n.LocalSnapComplete()
	}
}

func (n Node) RecvReject(port int) {
	n.unrelated = UnionOfSetInteger(port, n.unrelated)
	union := UnionOfSets(n.children, n.unrelated)
	var parentAsArray []int
	parentAsArray[0] = n.parent
	difference := Difference(n.neighbors, parentAsArray)
	sort.Ints(union)
	sort.Ints(difference)
	if reflect.DeepEqual(union, difference) {
		// node has received a MARKER on all (incoming) ports
		n.LocalSnapComplete()
	}
}

func UnionOfSetInteger(nonce int, colorset []int) []int {
	for _, elem := range colorset {
		if elem == nonce {
			return colorset
		}
	}
	return append(colorset, nonce)
}

func UnionOfSets(a, b []int) []int {
	m := make(map[int]bool)
	for _, item := range a {
		m[item] = true
	}
	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

func Difference(a, b []int) []int {
	mb := make(map[int]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []int
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (n Node) LocalSnapComplete() {
	fmt.Println("LOCAL_SNAP_COMPLETE")
	if len(n.children) == 0 {
		if n.parent != int(n.region_color) { //non-initiator leaf node
			SendInsweep(n.parent_port, n.neighborhood_color_set)
		} else { //initiator leaf node
			SendTerminate(n.region_color, n.self_port, n.neighborhood_color_set)
		}
	} else {
		n.AwaitInsweep()
		if n.parent != int(n.self_port) { //non-initiator leaf node
			SendInsweep(n.parent_port, n.neighborhood_color_set)
		} else { //initiator leaf node
			SendTerminate(n.region_color, n.self_port, n.neighborhood_color_set)
		}
	}
}

func RecordEmptyChannelState(port int) {
	//todo
}
func RecordChannelState(port int) {
	//todo
}

func main() {
	var node Node
	node.InitiateSnapshot()
}
