package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type WorkerNode struct {
	conn *grpc.ClientConn  // grpc client connection
	c    NodeServiceClient // grpc client
}

func (n *WorkerNode) Init() (err error) {
	// connect to master node
	n.conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return err
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
	fmt.Printf("Snapshot with id %s was create for volume %s\n",
		snap.GetSnapshotId(),
		v.GetVolumeId())
	return nil
}

func (n *WorkerNode) Start() {
	// log
	fmt.Println("worker node started")

	// report status
	_, _ = n.c.ReportStatus(context.Background(), &Request{})

	// assign task
	stream, _ := n.c.AssignTask(context.Background(), &Request{})
	for {
		// receive command from master node
		res, err := stream.Recv()
		if err != nil {
			return
		}

		// log command
		fmt.Println("received command: ", res.Data)

		// execute command
		parts := strings.Split(res.Data, " ")
		if err := exec.Command(parts[0], parts[1:]...).Run(); err != nil {
			fmt.Println(err)
		}
	}
}

var workerNode *WorkerNode

func GetWorkerNode() *WorkerNode {
	if workerNode == nil {
		// node
		workerNode = &WorkerNode{}

		// initialize node
		if err := workerNode.Init(); err != nil {
			panic(err)
		}
	}

	return workerNode
}
