package main

import (
	// (一部抜粋)
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	hellopb "mygrpc/pkg/grpc"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	scanner *bufio.Scanner
	client  hellopb.GreetingServiceClient
)

func main() {
	fmt.Println("start gRPC Client.")

	// 1. 標準入力から文字列を受け取るスキャナを用意
	scanner = bufio.NewScanner(os.Stdin)

	// 2. gRPCサーバーとのコネクションを確立
	address := "localhost:8080"
	conn, err := grpc.Dial(
		address,

		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal("Connection failed.")
		return
	}
	defer conn.Close()

	// 3. gRPCクライアントを生成
	client = hellopb.NewGreetingServiceClient(conn)

	for {
		fmt.Println("1: send Request")
		fmt.Println("2: HelloServerStream Method")
		fmt.Println("3: HelloClientStream Method")
		fmt.Println("4: exit")
		fmt.Print("please enter >")

		scanner.Scan()
		in := scanner.Text()

		switch in {
		case "1":
			Hello()
		case "2":
			HelloServerStream()
		case "3":
			HelloClientStream()
		case "4":
			fmt.Println("bye.")
			goto M
		}
	}
M:
}


// Unary RPCが引数を受け取り結果を返却するi/o
func Hello() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	// リクエストに使うHelloRequest型の生成
	req := &hellopb.HelloRequest{
		Name: name,
	}
	// Helloメソッドを実行
	res, err := client.Hello(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}

// Server Streaming RPCが引数を受け取り結果を返却するi/o
func HelloServerStream() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{
		Name: name,
	}

	// サーバから複数回レスポンスを受けるためのストリームを得る
	stream, err := client.HelloServerStream(context.Background(), req)

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		// ストリームからHelloResponse型レスポンスを得る
		// Recvメソッドでレスポンスを受け取るとき、これ以上受け取るレスポンスがないという状態なら、第一戻り値にはnil、第二戻り値のerrにはio.EOFが格納されています
		res, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			fmt.Println("all the responses have already received.")
			break
		}

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(res)
	}
}

// Client Streaming RPCが引数を受け取り結果を返却するi/o
func HelloClientStream() {
	// service(client)のHelloClientStreamメソッドの返り値から
	// 第一引数のGreetingService_HelloClientStreamClientインターフェースをstreamとして
	// 第二引数をerrとして受け取る
	stream, err := client.HelloClientStream(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	sendCount := 5
	fmt.Printf("Please enter %d names.\n", sendCount)
	
	// サーバに複数回リクエストを送信
	for i := 0; i < sendCount; i++ {
		scanner.Scan()
		name := scanner.Text()
		// GreetingService_HelloClientStreamClientインターフェースを表現するstreamからHelloRequest型の引数と共にSendメソッドを使用してリクエストを送る
		if err := stream.Send(&hellopb.HelloRequest{
			Name: name,
		}); err != nil {
			fmt.Println(err)
			return
		}
	}

	// リクエストを送信していたstreamのCloseAndRecvメソッドを呼び出すことでストリーム終端の伝達と、レスポンスを取得を行う
	res, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}

// ストリームで一連の通信を管理している．