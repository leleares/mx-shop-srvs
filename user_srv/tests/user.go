package main

import (
	"context"
	"fmt"
	"mx-shop-srvs/user_srv/proto"

	"google.golang.org/grpc"
)

var (
	userClient proto.UserClient
	conn       *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	userClient = proto.NewUserClient(conn)
}

func main() {
	TestCreateUser()
	conn.Close()
}

func TestGetUserList() {
	rsp, _ := userClient.GetUserList(context.Background(), &proto.PageInfo{Pn: 1, PSize: 10})

	for _, user := range rsp.Data {
		fmt.Println(user.Id, user.NickName, user.Password, user.Mobile)
		checkRes, _ := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          "generic password",
			EncryptedPassword: user.Password,
		})

		fmt.Println(checkRes.Success)
	}
}

func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, _ := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			NickName: fmt.Sprintf("rose%d", i),
			Mobile:   fmt.Sprintf("1663102162%d", i),
			Password: "admin123",
		})

		fmt.Println(rsp)
	}
}
