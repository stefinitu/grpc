# grpc
SD Project

Steps (HARD WAY ONLY FOR WINDOWS - CAN'T EXECUTE COMMANDS):
0. check system environment variables and path for your GOPATH (main.go, go.mod etc etc are in $GOPATH/SRC!!)
1. go get -v [all packages from files just like they are written in the go files] in $GOPATH/SRC
2. go install [all packages from files] in the same folder. If you have error, try go get without -v or go mod vendor 
3. Make sure Core folder is in $GOPATH/src (files from Core folder are in $GOPATH/src/core)
4. When you solve all the package errors: go mod vendor in $GOPATH/src !! IF YOU STILL HAVE PROBLEMS WITH PACKAGES/VENDOR: go run -mod=mod main.go master (or worker)
5. go run main.go master (If your $GOPATH is in Program Files or another C:// path, open Power Shell for Administrator!!)
6. docker run --rm --name sdk -d -p 9100:9100 -p 9110:9110 openstorage/mock-sdk-server  --THIS WILL START MOCK SDK SERVER instead of master grpc server from the code, ONLY FOR SNAPSHOT!
7. In another Terminal: go run main.go worker
9. If there is no error, congrats, you just created a Distributed System ++ Snapshot was created!

Steps (EASY WAY - WITH WSL UBUNTU 20.04 AND GOLANG 1.19.5):
1.sudo wget https://golang.org/dl/go1.19.5.linux-amd64.tar.gz
2. sudo tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz 
3. export PATH=$PATH:/usr/local/go/bin
4. cd [path to Documents/go/src] --> mine is /mnt/c/Users/[name]/Documents/Go/src
5. cp -r myproj /usr/local/go/src
6. cd /usr/local/go/src
7. go run main.go worker (master can be run from Powershell or WSL Terminal)
8.In another WSL TERMINAL : curl -X POST \
    -H "Content-Type=application/json" \
    -d '{"cmd": "touch /tmp/here"}' \
    http://localhost:9093/tasks
or Postman with Header [content from -H] and Body [content from -d] in JSON Format
9. worker received command!

PORTS:
MASTER - go run main.go master (9093 API)
MASTERWORKER - go run main.go masterworker 1 [ANYTHING] [NR OF PORTS TO DIAL] [ARRAY OF PORTS TO DIAL] (receive commands from API 9093), 
go run main.go masterworker 2 [PORT CREATED FOR LISTEN] [NR OF PORTS TO DIAL] [ARRAY OF PORTS TO DIAL] (API 9092)
WORKER - go run main.go worker (RECEIVE COMMANDS FROM 9092 API)

From: https://dev.to/tikazyq/golang-in-action-how-to-implement-a-simple-distributed-system-2n0n
