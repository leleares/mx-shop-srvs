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
	messageClient = proto.NewMessageClient(conn)
	userFavClient = proto.NewUserFavClient(conn)
}

func main() {
	// TestCreateAddress()
	// TestUpdateAddress()
	// TestAddressList()
	// TestDeleteAddress()
	// TestCreateMessage()
	// TestMessageList()
	// TestAddUserFav()
	// TestGetUserFavDetail()
	// TestGetFavList()
	TestDelUserFav()
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

func TestCreateMessage() {
	resp, err := messageClient.CreateMessage(context.Background(), &proto.MessageRequest{
		UserId:      10,
		MessageType: 1,
		Subject:     "主题就是留言",
		Message:     "这是留言balabala",
		File:        "https://piccdn2.umiwi.com/fe-oss/default/MTc3Mzg5MTAyMDg4.png",
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestMessageList() {
	resp, err := messageClient.MessageList(context.Background(), &proto.MessageRequest{
		UserId: 10,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestAddUserFav() {
	resp, err := userFavClient.AddUserFav(context.Background(), &proto.UserFavRequest{
		UserId:  10,
		GoodsId: 421,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestGetUserFavDetail() {
	resp, err := userFavClient.GetUserFavDetail(context.Background(), &proto.UserFavRequest{
		UserId:  10,
		GoodsId: 421,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestGetFavList() {
	resp, err := userFavClient.GetFavList(context.Background(), &proto.UserFavRequest{
		UserId: 10,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestDelUserFav() {
	resp, err := userFavClient.DeleteUserFav(context.Background(), &proto.UserFavRequest{
		UserId:  10,
		GoodsId: 421,
	})

	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}
