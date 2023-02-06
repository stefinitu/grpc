package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

func (n *WorkerNode) Init(port string) (err error) {

	n.conn, err = grpc.Dial("localhost:"+port, grpc.WithInsecure()) //dial la nodul Master

	if err != nil {
		return err
	}
	n.c = NewNodeServiceClient(n.conn)
	return nil
}

func (n *WorkerNode) Start(port string) {
	fmt.Println("worker node started")

	_, _ = n.c.ReportStatus(context.Background(), &Request{})

	// assign task
	stream, _ := n.c.AssignTask(context.Background(), &Request{})
	sent := false
	for {
		// receive command from master node
		res, err := stream.Recv() //asteapta primirea comenzilor/mesajelor-marker de la nodul Master
		if err != nil {
			return
		}

		// log command
		fmt.Println("received command/marker: ", res.Data)

		// execute command
		parts := strings.Split(res.Data, " ")
		numeric := regexp.MustCompile(`\d`).MatchString(parts[0])

		if numeric == false { //daca nu primeste numar (Nonce), atunci a primit comanda
			mycom := exec.Command(parts[0], parts[1:]...)
			if err := mycom.Run(); err != nil {
				fmt.Println(err)
				output, _ := mycom.CombinedOutput()
				fmt.Println(string(output))

			}
		}
		if numeric == true {
			fmt.Println("Received Marker: ", parts[0]) //a primit mesaj-marker

			if sent == false {
				// se conecteaza la serverul din container pentru Snapshot
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
				volumes := api.NewOpenStorageVolumeClient(n.conn) //creeaza volum
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

				snap, err := volumes.SnapshotCreate( //Snapshot
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

				fmt.Println("LOCAL SNAP COMPLETED!")                            //fiind nod-frunza, se instiinteaza LOCAL SNAP COMPLETE
				n.conn, err = grpc.Dial("localhost:"+port, grpc.WithInsecure()) //reconectarea la nodul Master pentru noi comenzi
				if err != nil {
					os.Exit(1)
				}
				fmt.Println("Connected to " + port)
			}
			if sent == true {
				fmt.Println("Marker already received! REJECT!") //daca a primit din nou un Marker, trimite mesajul REJECT
				// n.conn, err = grpc.Dial("localhost:"+port, grpc.WithInsecure())
				// if err != nil {
				// 	os.Exit(1)
				// }
				// fmt.Println("Connected to " + port)

				// n.c = NewNodeServiceClient(n.conn)
			}
			if sent == false {
				sent = true
			}
		}
	}
}

var workerNode *WorkerNode

func GetWorkerNode(port string) *WorkerNode {
	if workerNode == nil {

		workerNode = &WorkerNode{}
		if err := workerNode.Init(port); err != nil {
			panic(err)
		}
	}

	return workerNode
}
