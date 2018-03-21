package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sbramin/grpc-demo/pkg/pb/example"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial("localhost:8090", grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
	if err != nil {
		log.Fatal("conn err", err)
	}
	cli := example.NewExampleClient(conn)
	defer conn.Close()

	resp, err := cli.GetExample(ctx, &example.Request{Req: "Bob"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)

}
