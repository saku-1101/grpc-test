package main

import (
	// (一部抜粋)
	"context"
	"fmt"
	"log"
	hellopb "mygrpc/pkg/grpc"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// 自作サービス構造体の定義
type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

// 「HelloRequest型のリクエストを受け取って、HelloResponse型のレスポンスを返す」Helloメソッド
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	// リクエストからnameフィールドを取り出して
	// "Hello, [名前]!"というレスポンスを返す
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
// 👆Helloメソッドを実装した自作サービス構造体myServer型の定義完成

// 👇myServer型を提供する処理の実装
func NewMyServer() *myServer {
	return &myServer{}
}


func main() {
	// 1. 8080番portのLisnterを作成
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// 2. gRPCサーバーを作成
	s := grpc.NewServer()

	// 3. gRPCサーバーにGreetingServiceを登録
	hellopb.RegisterGreetingServiceServer(s, NewMyServer())

	// 4. サーバリフレクション設定
	reflection.Register(s)
	// 5. 作成したgRPCサーバーを、8080番ポートで稼働させる
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// 6.Ctrl+Cが入力されたらGraceful shutdownされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
