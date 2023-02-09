package core

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Master2Node is the node instance
type Master2Node struct {
	conn    *grpc.ClientConn       // grpc client connection
	c       NodeServiceClient      // grpc client
	api     *gin.Engine            // api server
	ln      net.Listener           // listener
	svr     *grpc.Server           // grpc server
	nodeSvr *NodeServiceGrpcServer // node service
}

var nodeMaster2 *Master2Node

// GetMasterNode returns the node instance
func GetMaster2Node(port string) *Master2Node {
	if nodeMaster2 == nil {
		// node
		nodeMaster2 = &Master2Node{}

		if err := nodeMaster2.Init(port); err != nil {
			panic(err)
		}
	}

	return nodeMaster2
}

func (n *Master2Node) Init(port string) (err error) {
	//nodul-radacina inregistreaza Snapshot odata ce este rulat
	n.conn, err = grpc.Dial("localhost:9100", grpc.WithInsecure())
	if err != nil {
		os.Exit(1)

	}

	// grpc client
	n.c = NewNodeServiceClient(n.conn)
	cluster := api.NewOpenStorageClusterClient(n.conn)
	clusterInfo, err := cluster.InspectCurrent(
		context.Background(),
		&api.SdkClusterInspectCurrentRequest{})
	if err != nil {
		gerr, _ := status.FromError(err)
		fmt.Printf("Error Code[%d] Message[%s]\n",
			gerr.Code(), gerr.Message())
		os.Exit(1)
	}
	fmt.Printf("Connected to Cluster %s\n",
		clusterInfo.GetCluster().GetId())
	volumes := api.NewOpenStorageVolumeClient(n.conn)
	v, err := volumes.Create(
		context.Background(),
		&api.SdkVolumeCreateRequest{
			Name: "myvol",
			Spec: &api.VolumeSpec{
				Size:    100 * 1024 * 1024 * 1024,
				HaLevel: 3,
			},
		})
	if err != nil {
		gerr, _ := status.FromError(err)
		fmt.Printf("Error Code[%d] Message[%s]\n",
			gerr.Code(), gerr.Message())
		os.Exit(1)
	}
	fmt.Printf("Volume 100Gi created with id %s\n", v.GetVolumeId())

	snap, err := volumes.SnapshotCreate(
		context.Background(),
		&api.SdkVolumeSnapshotCreateRequest{
			VolumeId: v.GetVolumeId(),
			Name:     fmt.Sprintf("snap-%v", time.Now().Unix()),
		},
	)
	if err != nil {
		gerr, _ := status.FromError(err)
		fmt.Printf("Error Code[%d] Message[%s]\n",
			gerr.Code(), gerr.Message())
		os.Exit(1)
	}
	fmt.Printf("Snapshot with id %s was created for volume %s\n",
		snap.GetSnapshotId(),
		v.GetVolumeId())

	n.ln, err = net.Listen("tcp", ":"+port) //asculta pe un port primit ca argument al comenzii de rulare
	if err != nil {
		return err
	}

	// grpc server
	n.svr = grpc.NewServer()

	n.nodeSvr = GetNodeServiceGrpcServer()

	// register node service to grpc server
	RegisterNodeServiceServer(nodeMaster2.svr, n.nodeSvr)

	//NONCE
	rand.Seed(time.Now().UnixNano())
	var Nonce = rand.Uint64() //Nonce generated (128-bit uint is hard to generate)
	var stringNonce = fmt.Sprint(Nonce)
	//write its Nonce to Universal Color Set
	f, err := os.OpenFile("./universal_color_set.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(stringNonce + "\n"); err != nil {
		panic(err)
	}

	// api
	n.api = gin.Default()
	n.api.POST("/tasks", func(c *gin.Context) {
		var payload struct {
			Cmd string `json:"cmd"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// send command to node service
		fmt.Println(payload.Cmd)
		n.nodeSvr.CmdChannel <- payload.Cmd

		c.AbortWithStatus(http.StatusOK)
	})

	n.api.POST("/marker", func(c *gin.Context) {
		var payload struct {
			Marker string `json:"marker"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// send command to node service
		fmt.Println(payload.Marker)
		n.nodeSvr.CmdChannel <- payload.Marker

		c.AbortWithStatus(http.StatusOK)

	})

	return nil
}

func (n *Master2Node) Start(port string) {
	// start grpc server
	go n.svr.Serve(n.ln)

	// start api server
	_ = n.api.Run(":9094")

	n.svr.Stop()

}
