package main

import (
	"github.com/golang/protobuf/ptypes"
	pb "gogrpc/chat"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"fmt"
)

type server struct {
	broadcast chan *pb.ChatMessage
	clients []chan *pb.ChatMessage
}



func (s *server) Chat(stream pb.ChatService_ChatServer) error {
	go s.sendBroadcast(stream)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		fmt.Printf("(%s) %s : %s \n", ptypes.TimestampString(in.Time), in.User, in.Message)

		s.broadcast <- in
	}

}

func (s *server) sendBroadcast(stream pb.ChatService_ChatServer) {
	clientChan := make(chan *pb.ChatMessage, 100)
	s.clients = append(s.clients, clientChan)
	for {
		msg := <- clientChan
		err := stream.Send(msg)
		if err!=nil {
			fmt.Println(err)
		}
	}
}

func (s *server) BroadcastBlaster() {
	for res := range s.broadcast {
		for _, client := range s.clients {
			client <- res
		}
	}
}

func main() {
	serverAddress := "localhost:8090"
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("server listen at ", serverAddress)

	grpcServer := grpc.NewServer()
	s := &server{
		broadcast: make(chan *pb.ChatMessage, 1000),
	}

	pb.RegisterChatServiceServer(grpcServer, s)

	go s.BroadcastBlaster()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}