# grpc
SD Project
I hate this. :(

Steps:
1. go get -v <all packages from files>
2. go install <all packages from files>
3. Make sure your project is in $GOPATH/src (check system environment variables for your path) (main.go, go.mod etc etc is in SRC!!)
4. Make sure Core folder is in $GOPATH/src (files from Core folder are in $GOPATH/src/core)
5. When you solve all the package errors: go mod vendor in $GOPATH/src !!
6. go run main.go master (If your $GOPATH is in Program Files or another C:// path, open Power Shell for Administrator!!
7. In another Terminal go run main.go worker
8. If there is no error, congrats, you just created a Distributed System!
9. run docker container for mock server with this command:
