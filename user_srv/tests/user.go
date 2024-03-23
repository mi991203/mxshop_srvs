package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"mxshop_srvs/user_srv/proto"
)

func init() {
	conn, _ = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	userClient = proto.NewUserClient(conn)
}

var (
	conn       *grpc.ClientConn
	userClient proto.UserClient
)

func main() {
	defer conn.Close()
	TestCreateUser()
	TestGetUserList()
}

func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, err := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			NickName: fmt.Sprintf("Jack%d", i),
			Mobile:   fmt.Sprintf("1111111111%d", i),
			Password: "admin123",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("用户创建成功%s", rsp.String()))
	}
}

func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 5,
	})
	if err != nil {
		panic(err)
	}
	for _, user := range rsp.Data {
		fmt.Println(user.Mobile, user.NickName, user.Password)
		checkRsp, err := userClient.CheckPassword(context.Background(), &proto.CheckPasswordInfo{
			Password:          "admin123",
			EncryptedPassword: user.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkRsp.Success)
	}
}
