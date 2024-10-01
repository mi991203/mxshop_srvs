package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/handler"
	"mxshop_srvs/user_srv/initialize"
	"mxshop_srvs/user_srv/proto"
	"net"
)

func main() {
	// 初始化
	initialize.InitConfig()
	initialize.InitLogger()
	initialize.InitDB()

	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")

	flag.Parse()

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("fialed to listen:" + err.Error())
	}

	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 服务注册
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"

	client, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}

	// 添加健康检查
	check := &api.AgentServiceCheck{
		GRPC:                           "192.168.0.109:50051",
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}

	registration := new(api.AgentServiceRegistration)
	registration.Address = "192.168.0.109"
	registration.Port = *Port
	registration.Name = global.ServerConfig.Name
	registration.Tags = []string{"mxshop"}
	registration.ID = global.ServerConfig.Name
	registration.Check = check

	if err = client.Agent().ServiceRegister(registration); err != nil {
		panic(err)
	}

	if err := server.Serve(lis); err != nil {
		panic("fail to start grpc:" + err.Error())
	}

}
