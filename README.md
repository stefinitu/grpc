# grpc
SD Project
I hate this. :(

Steps:
0. check system environment variables and path for your GOPATH (main.go, go.mod etc etc are in $GOPATH/SRC!!)
1. go get -v <all packages from files> in $GOPATH/SRC
2. go install <all packages from files> in the same folder
3. Make sure Core folder is in $GOPATH/src (files from Core folder are in $GOPATH/src/core)
4. When you solve all the package errors: go mod vendor in $GOPATH/src !!
5. go run main.go master (If your $GOPATH is in Program Files or another C:// path, open Power Shell for Administrator!!) --DON'T DO THIS IF YOU WANT SNAPSHOT!
6. docker run --rm --name sdk -d -p 9100:9100 -p 9110:9110 openstorage/mock-sdk-server  --THIS WILL START MOCK SDK SERVER instead of master grpc server from the code, ONLY FOR SNAPSHOT!
7. In another Terminal: go run main.go worker AFTER YOU CHANGE IN worker_node.go the port to 9100 ---> grpc.Dial(localhost:9100) for SNAPSHOT CASE!! (OTHERWISE, JUST KEEP IT THE WAY IT IS)
8. If there is no error, congrats, you just created a Distributed System ++ Snapshot was created!
