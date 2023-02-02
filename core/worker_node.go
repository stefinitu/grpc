package core

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
	// n.conn, err = grpc.Dial("localhost:9100", grpc.WithInsecure())
	// if err != nil {
	// 	return err

	// }

	// // grpc client
	// n.c = NewNodeServiceClient(n.conn)
	// cluster := api.NewOpenStorageClusterClient(n.conn)
	// clusterInfo, err := cluster.InspectCurrent(
	// 	context.Background(),
	// 	&api.SdkClusterInspectCurrentRequest{})
	// if err != nil {
	// 	gerr, _ := status.FromError(err)
	// 	fmt.Printf("Error Code[%d] Message[%s]\n",
	// 		gerr.Code(), gerr.Message())
	// 	os.Exit(1)
	// }
	// fmt.Printf("Connected to Cluster %s\n",
	// 	clusterInfo.GetCluster().GetId())
	// volumes := api.NewOpenStorageVolumeClient(n.conn)
	// v, err := volumes.Create(
	// 	context.Background(),
	// 	&api.SdkVolumeCreateRequest{
	// 		Name: "myvol",
	// 		Spec: &api.VolumeSpec{
	// 			Size:    100 * 1024 * 1024 * 1024,
	// 			HaLevel: 3,
	// 		},
	// 	})
	// if err != nil {
	// 	gerr, _ := status.FromError(err)
	// 	fmt.Printf("Error Code[%d] Message[%s]\n",
	// 		gerr.Code(), gerr.Message())
	// 	os.Exit(1)
	// }
	// fmt.Printf("Volume 100Gi created with id %s\n", v.GetVolumeId())

	// snap, err := volumes.SnapshotCreate(
	// 	context.Background(),
	// 	&api.SdkVolumeSnapshotCreateRequest{
	// 		VolumeId: v.GetVolumeId(),
	// 		Name:     fmt.Sprintf("snap-%v", time.Now().Unix()),
	// 	},
	// )
	// if err != nil {
	// 	gerr, _ := status.FromError(err)
	// 	fmt.Printf("Error Code[%d] Message[%s]\n",
	// 		gerr.Code(), gerr.Message())
	// 	os.Exit(1)
	// }
	// fmt.Printf("Snapshot with id %s was create for volume %s\n",
	// 	snap.GetSnapshotId(),
	// 	v.GetVolumeId())

	// fmt.Println(snap.Descriptor())
	// fmt.Println("\n\n")

	n.conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return err
	}
	n.c = NewNodeServiceClient(n.conn)
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
		numeric := regexp.MustCompile(`\d`).MatchString(parts[0])

		if numeric == false {
			mycom := exec.Command(parts[0], parts[1:]...)
			if err := mycom.Run(); err != nil {
				fmt.Println(err)
				output, _ := mycom.CombinedOutput()
				fmt.Println(string(output))

			}
		}
		if numeric == true {
			fmt.Println("Received Marker: ", parts[0])
			file, _ := os.Open("./universal_color_set.txt")
			fileScanner := bufio.NewScanner(file)
			lineCount := 0
			for fileScanner.Scan() {
				lineCount++
			}
			if lineCount == 2 {
				// connect to master node fro SNAPSHOT
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

				//NONCE
				rand.Seed(time.Now().UnixNano())
				var Nonce3 = rand.Int() //Nonce generated
				//write Nonce to Universal Color Set
				f, err := os.OpenFile("./universal_color_set.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				if _, err = f.WriteString(strconv.Itoa(Nonce3) + "\n"); err != nil {
					panic(err)
				}
				fmt.Println("Wrote in file")
				fmt.Println("LOCAL SNAP COMPLETED!")
				n.conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure())
				if err != nil {
					os.Exit(1)
				}
				fmt.Println("Connected to 50051")
			}
			if lineCount == 3 {
				fmt.Println("Nonce already generated! REJECT!")
				n.conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure())
				if err != nil {
					os.Exit(1)
				}
				fmt.Println("Connected to 50051")

				n.c = NewNodeServiceClient(n.conn)
			}
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
