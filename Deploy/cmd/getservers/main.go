package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
)

func main() {
	addr := flag.String("addr", ":8400", "service address")
	flag.Parse()

	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := api.NewLogClient(conn)
	ctx := context.Background()

	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Servers:")
	for _, server := range res.Servers {
		fmt.Printf("\t- %v\n", server)
	}

	// --- Used client (produce and consume) to test the connection ---

	produceRes, err := client.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("Hello, World!"),
		},
	})
	if err != nil {
		log.Fatalf("Produce error: %v", err)
	}
	fmt.Printf("Produced record at offset %d\n", produceRes.Offset)

	consumeRes, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produceRes.Offset,
	})
	if err != nil {
		log.Fatalf("Consume error: %v", err)
	}
	fmt.Printf("Consumed record: %s\n", string(consumeRes.Record.Value))
}
