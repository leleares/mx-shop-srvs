package main

import (
	"google.golang.org/grpc"
)

var (
	// orderClient proto.InventoryClient
	conn *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	// inventoryClient = proto.NewInventoryClient(conn)
}

func main() {
}
