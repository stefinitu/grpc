# grpc
SD Project

Steps (EASY WAY - WITH WSL UBUNTU 20.04 AND GOLANG 1.19.5):
1.sudo wget https://golang.org/dl/go1.19.5.linux-amd64.tar.gz
2. sudo tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz 
3. export PATH=$PATH:/usr/local/go/bin
4. cd [path to Documents/go/src] --> mine is /mnt/c/Users/[name]/Documents/Go/src
5. cp -r myproj /usr/local/go/src
6. cd /usr/local/go/src
7. docker run --rm --name sdk -d -p 9100:9100 -p 9110:9110 openstorage/mock-sdk-server  --THIS WILL START MOCK SDK SERVER
8. go run main.go worker (master can be run from Powershell or WSL Terminal)
9.In another WSL TERMINAL :

curl -X POST \
    -H "Content-Type=application/json" \
    -d '{"cmd": "touch /tmp/here"}' \
    http://localhost:9093/tasks
    (FOR LINUX COMMANDS)
       
    curl -X POST \
    -H "Content-Type=application/json" \
    -d '{"marker": "7286039776793402467"}' \
    http://localhost:9093/marker  
    (FOR MARKER)
or Postman with Header [content from -H] and Body [content from -d] in JSON Format
10. worker received command!

PORTS:
MASTER - go run main.go master (9093 API)
MASTERWORKER - go run main.go masterworker 1 [ANYTHING] [NR OF PORTS TO DIAL] [ARRAY OF PORTS TO DIAL] (receive commands from API 9093), 
go run main.go masterworker 2 [PORT CREATED FOR LISTEN] [NR OF PORTS TO DIAL] [ARRAY OF PORTS TO DIAL] (API 9092)
WORKER - go run main.go worker (RECEIVE COMMANDS FROM 9092 API)

implem.go -> Algorithm implementation in Golang language

From: https://dev.to/tikazyq/golang-in-action-how-to-implement-a-simple-distributed-system-2n0n
