package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial("localhost:8091", grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
	if err != nil {
		log.Fatal("conn err", err)
	}
	cli := tps.NewThirdPartyServiceClient(conn)
	defer conn.Close()

	id := int64(1)

	resp, err := cli.Echo(ctx, &tps.Input{Id: id, SuperMessage: "Bob"})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			id++
			resp, err = cli.Echo(ctx, &tps.Input{Id: id, SuperMessage: "Bob"})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	fmt.Println(resp)

}
