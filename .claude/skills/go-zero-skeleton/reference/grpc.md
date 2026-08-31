# GRPC 服务指南

骨架不依赖 goctl 的 zRPC 体系，直接使用原生 grpc 库，proto 产物与服务端代码均放在 `app/rpc/` 下。

## 新增 GRPC 服务端

### 1. 编写 proto 文件（置于项目根目录）

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
}
```

约定：每个服务一个 proto 文件，`go_package` 指向 `./app/rpc/<包名>`。

### 2. 生成代码

```bash
protoc --go_out=. --go-grpc_out=. user-api.proto
```

产物生成到 `app/rpc/user_api/`。

### 3. 编写服务端 `app/rpc/server.go`

```go
package rpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/limingxinleo/go-zero-skeleton/app/config"
	pb "github.com/limingxinleo/go-zero-skeleton/app/rpc/user_api"
	"google.golang.org/grpc"
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &UserServiceServer{})

	fmt.Printf("GRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

业务逻辑较重时，建议将实现委托给 `app/service` 中的服务，rpc 层只做参数转换。

### 4. 在入口挂载

`main.go` 中 `controller.RegisterHandlers(...)` 之后、`server.Start()` 之前追加：

```go
go rpc.StartGRPCServer(app.GetApplication().Config)
```

### 5. 增加端口配置

`app/config/config.go` 的 `Config` 增加 `GrpcPort int`，并在 `etc/main-api.yaml`、`etc/unit-api.yaml` 中添加 `GrpcPort: 9600`。

## GRPC 客户端（调用外部服务）

在对应 proto 包（如 `app/rpc/user_api`）中编写连接构造代码：

```go
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

使用建议：
- `*grpc.ClientConn` 创建成本较高，应缓存复用（如挂在 `ServiceContext` 中初始化一次），不要每次请求都建连。
- 目标地址写入 `etc/*.yaml` 配置，不要硬编码。
- 调用示例：

```go
conn, err := user_api.NewUserServiceConn(host, port)
if err != nil { ... }
defer conn.Close()

client := user_api.NewUserServiceClient(conn)
resp, err := client.GetChildren(ctx, &user_api.UserIdRequest{Id: 1})
```
