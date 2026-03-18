package main

import (
	"mx-shop-srvs/user_interaction_srv/proto"

	"google.golang.org/grpc"
)

var (
	addressClient proto.AddressClient
	messageClient proto.MessageClient
	userFavClient proto.UserFavClient
	conn          *grpc.ClientConn
)

// func init() {
// 	var err error
// 	conn, err = grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
// 	if err != nil {
// 		panic(err)
// 	}

// 	inventoryClient = proto.NewInventoryClient(conn)
// }

func main() {
}
