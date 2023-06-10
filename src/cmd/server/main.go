package main

import (
	// (一部抜粋)
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	hellopb "mygrpc/pkg/grpc"
	"net"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// 自作サービス構造体の定義
type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

// 「HelloRequest型のリクエストを受け取りHelloResponse型のレスポンスを返す」ロジック
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	// リクエストからnameフィールドを取り出して
	// "Hello, [名前]!"というレスポンスを返す
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
// 👆Helloメソッドを実装した自作サービス構造体myServer型の定義完成

// Server Streaming RPCでリクエストを受け取りレスポンスを返すロジック
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
	resCount := 5
	for i := 0; i < resCount; i++ {
		// Sendメソッドを何度も実行することで何度もクライアントにレスポンスを返すことができる
		// サーバからのストリーミングを実現してる
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	// returnでメソッドをおわらせる＝ストリームの終端
	return nil
}

// Client Streaming RPCでリクエストを受け取りレスポンスを返すロジック
func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
	nameList := make([]string, 0)
	for {
		// streamのRecvメソッドを読んでリクエスト内容を受け取る
		// これを何度も呼ぶことにより，クライアントから複数回送られてくるリクエスト内容を受け取る
		req, err := stream.Recv()
		
		// 全部受け取った後の処理
		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			// SendAndCloseを呼ぶことでレスポンスを返す
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}
		if err != nil {
			return err
		}
		nameList = append(nameList, req.GetName())
	}
}

// Bidirectional streaming RPC
func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error {
	for {
		// リクエスト受信
		req, err := stream.Recv()

		// 得られたエラーがio.EOFならばもうリクエストは送られてこないのでnilで処理終了
		if errors.Is(err, io.EOF) {
			return nil
		}

		// エラーの場合はそのままエラーを返して終了
		if err != nil {
			return err
		}
		
		// 受信したリクエストから名前を取得して応答メッセージを作成
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		// 応答メッセージをクライアントに送信
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil { //送信中にエラーが発生した場合はそのままエラーを返して終了
			return err
		}

		// クライアントはストリームを介して複数のリクエストを送信し, サーバーは各リクエストに対して個別の応答を返すことができる
	}
}

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
	// gRPCの通信はProtocol Bufferでシリアライズされている
	// rRPCurlに「gRPCサーバーそのものから、protoファイルの情報を取得する」ことで「シリアライズのルール」を知らせるため
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
