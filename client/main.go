package main

import (
	"github.com/golang/protobuf/ptypes"
	"fmt"
	"context"
	"time"
	"google.golang.org/grpc"
	"io"
	"log"
	pb "gogrpc/chat"
	"os"
)

func main() {
	address := "localhost:8090"
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewChatServiceClient(conn)

	name := "anonimus"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := c.Chat(ctx)
	if err != nil {
		log.Fatal("chat error : %v", err)
	}

	msg := &pb.ChatMessage{
		User:                 name,
		Message:              name + " join the chat",
		Time:                 ptypes.TimestampNow(),
	}
	stream.Send(msg)

	go func() {
		ctr := 1
		for {
			msg := &pb.ChatMessage{
				User:                 name,
				Message:              fmt.Sprintf("%s counting %d", name, ctr),
				Time:                 ptypes.TimestampNow(),
			}
			stream.Send(msg)
			ctr++
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Fatalf("error eof : %v", err)
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		fmt.Printf("(%s) %s : %s \n", ptypes.TimestampString(in.Time), in.User, in.Message)

	}
}
