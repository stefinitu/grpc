package core

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MasterNode is the node instance
type MasterWorkerNode struct {
	conn    *grpc.ClientConn       // grpc client connection
	c       NodeServiceClient      // grpc client
	api     *gin.Engine            // api server
	ln      net.Listener           // listener
	svr     *grpc.Server           // grpc server
	nodeSvr *NodeServiceGrpcServer // node service
}

var nodemw *MasterWorkerNode

// GetMasterNode returns the node instance
func GetMasterWorkerNode(nr string, port string, nrDial string, portDials []string) *MasterWorkerNode {
	if nodemw == nil {
		// node
		nodemw = &MasterWorkerNode{}

		// initialize node
		if err := nodemw.Init(nr, port, nrDial, portDials); err != nil {
			panic(err)
		}
	}

	return nodemw
}

func (n *MasterWorkerNode) Init(nr string, port string, nrDial string, portDials []string) (err error) {
	// // connect to master node
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
	// fmt.Printf("Snapshot with id %s was created for volume %s\n",
	// 	snap.GetSnapshotId(),
	// 	v.GetVolumeId())
	nrDialsInt, _ := strconv.Atoi(nrDial)
	for i := 0; i < nrDialsInt; i++ {
		n.conn, err = grpc.Dial("localhost:"+portDials[i], grpc.WithInsecure())
		if err != nil {
			return err
		}
		fmt.Println("Connected to " + portDials[i])

		n.c = NewNodeServiceClient(n.conn)
	}
	if nr == "2" {
		// grpc server listener with port as 50051
		n.ln, err = net.Listen("tcp", ":"+port)
		if err != nil {
			return err
		}

		// grpc server
		n.svr = grpc.NewServer()

		// node service
		n.nodeSvr = GetNodeServiceGrpcServer()

		// register node service to grpc server
		RegisterNodeServiceServer(nodemw.svr, n.nodeSvr)

		// api
		n.api = gin.Default()
		n.api.POST("/tasks", func(c *gin.Context) {
			// parse payload
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
			// parse payload
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
		n.api.POST("/insweep", func(c *gin.Context) {
			// parse payload
			var payload struct {
				Marker string `json:"insweep"`
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
	}
	return nil
}

func (n *MasterWorkerNode) Start(nr string, port string, nrDial string, portDials []string) {
	if nr == "1" {
		//commands from master
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

			//Check if node received a command or a Marker
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
				if lineCount == 1 {
					fmt.Println("ACCEPT! SNAPSHOT/NONCE WILL BE GENERATED!")
					// connect to master node from SNAPSHOT
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

					// //NONCE
					// rand.Seed(time.Now().UnixNano())
					// var Nonce = rand.Uint64() //Nonce generated (128-bit uint it is hard to generate)
					// var stringNonce = fmt.Sprint(Nonce)

					// //write Nonce to Universal Color Set
					// f, err := os.OpenFile("./universal_color_set.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
					// if err != nil {
					// 	panic(err)
					// }

					// defer f.Close()

					// if _, err = f.WriteString(stringNonce + "\n"); err != nil {
					// 	panic(err)
					// }
					// fmt.Println("Wrote in file")

					nrDialsInt, _ := strconv.Atoi(nrDial)
					for i := 0; i < nrDialsInt; i++ {
						n.conn, err = grpc.Dial("localhost:"+portDials[i], grpc.WithInsecure())
						if err != nil {
							os.Exit(1)
						}
						fmt.Println("Connected to " + portDials[i])

						n.c = NewNodeServiceClient(n.conn)
					}
				}
				if lineCount == 2 {
					fmt.Println("Nonce already generated! REJECT!")
					nrDialsInt, _ := strconv.Atoi(nrDial)
					for i := 0; i < nrDialsInt; i++ {
						n.conn, err = grpc.Dial("localhost:"+portDials[i], grpc.WithInsecure())
						if err != nil {
							os.Exit(1)
						}
						fmt.Println("Connected to " + portDials[i])

						n.c = NewNodeServiceClient(n.conn)
					}
				}
			}
		}
	}
	if nr == "2" {

		// start grpc server
		go n.svr.Serve(n.ln)

		// start api server
		_ = n.api.Run(":9092")

		// wait for exit
		n.svr.Stop()
	}
}
