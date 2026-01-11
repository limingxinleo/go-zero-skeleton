# GoZero Skeleton

## 数据库

### 生成模型

1. 使用原生组件

将数据库 `ddl` 语句放到 `app/model` 中，执行下述命令生成所有模型

```shell
goctl model mysql ddl --src ./app/model/ddl/mysql.sql --dir ./app/model
```

2. 使用 Gorm

```shell
go run cmd/main.go gen:mode
```

## RPC

### GRPC

1. 编写 proto 文件，置于根目录中，如下

```protobuf
syntax = "proto3";

package user_api;
option go_package = "./app/rpc/user_api;user_api";

service UserService {
  rpc GetChildren (UserIdRequest) returns (ChildrenResponse);
}

message UserIdRequest {
  uint64 id = 1;
}

message ChildrenResponse {
  repeated ChildSchema children = 1;
}

message ChildSchema {
  uint64 id = 1;
  string nick_name = 2;
  uint64 gender = 3;
  string birthday = 4;
  string avatar = 5;
}
```

2. 生成代码

```shell
protoc --go_out=. --go-grpc_out=. user-api.proto
```

3. 增加 `app/rpc/server.go` 文件。

```
package rpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"github.com/limingxinleo/go-zero-skeleton/app/config"
	pb "github.com/limingxinleo/go-zero-skeleton/app/rpc/user_api"
	"net"
)

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetChildren(ctx context.Context, req *pb.UserIdRequest) (*pb.ChildrenResponse, error) {
	var result []*pb.ChildSchema
	// 填充数据
	return &pb.ChildrenResponse{Children: result}, nil
}

func StartGRPCServer(conf *config.Config) {
	// 监听 TCP 端口
	// GrpcPort 可以写死，也可以写到配置中，如何编写配置，这里不做介绍
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &UserServiceServer{})

	// 启动服务器并监听传入的连接
	fmt.Printf("GRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

```

4. 修改入口文件 `main.go`

```
package main

import (
	"fmt"
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/controller"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	logx.MustSetup(logx.LogConf{
		ServiceName: app.GetApplication().Config.Name,
		Level:       "info",
		TimeFormat:  "2006-01-02 15:04:05.000",
	})

	server := rest.MustNewServer(
		app.GetApplication().Config.RestConf,
		rest.WithCustomCors(
			func(header http.Header) {
				header.Set("Access-Control-Allow-Headers", "DNT,Keep-Alive,User-Agent,Cache-Control,Content-Type,Authorization")
			},
			nil,
			"*",
		),
	)
	defer server.Stop()

	server.Use(kernel.ServerMiddleware)

	controller.RegisterHandlers(server, app.GetApplication().ServiceContext)
	
	// 增加如下代码
	go rpc.StartGRPCServer(app.GetApplication().Config)

	fmt.Printf("Starting server at %s:%d...\n", app.GetApplication().Config.Host, app.GetApplication().Config.Port)
	server.Start()
}
```

5. 增加调用代码

```
package user_api

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 此方法创建的 Conn 可以进行缓存
func NewUserServiceConn(host string, port int) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

```
