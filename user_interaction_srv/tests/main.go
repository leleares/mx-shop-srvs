package main

import (
	"context"
	"mx-shop-srvs/user_interaction_srv/model"
	"mx-shop-srvs/user_interaction_srv/proto"

	"google.golang.org/grpc"
)

var (
	addressClient proto.AddressClient
	messageClient proto.MessageClient
	userFavClient proto.UserFavClient
	conn          *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50054", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	addressClient = proto.NewAddressClient(conn)
}

func main() {
	// TestCreateAddress()
	// TestUpdateAddress()
	// TestAddressList()
	// TestDeleteAddress()
}

func TestCreateAddress() {
	resp, err := addressClient.CreateAddress(context.Background(), &proto.AddressRequest{
		UserId:       10,
		Province:     "河北省",
		City:         "邯郸市",
		District:     "曲周县",
		Address:      "候村镇龙李庄村105号",
		SignerName:   "李哈哈",
		SignerMobile: "16632021639",
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestUpdateAddress() {
	resp, err := addressClient.UpdateAddress(context.Background(), &proto.AddressRequest{
		Id:           1,
		UserId:       10,
		Province:     "河北省",
		City:         "邯郸市",
		District:     "曲周县",
		Address:      "安寨镇龙李庄村105号",
		SignerName:   "李哈哈",
		SignerMobile: "16632021639",
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestAddressList() {
	resp, err := addressClient.GetAddressList(context.Background(), &proto.AddressRequest{
		UserId: 10,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestDeleteAddress() {
	resp, err := addressClient.DeleteAddress(context.Background(), &proto.AddressRequest{
		Id:     1,
		UserId: 10,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}
