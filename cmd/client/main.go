package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/slayerbirden/chat/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var uid string

func main() {
	conn, err := grpc.Dial("localhost:9090", grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(time.Duration(1*time.Second)))
	if err != nil {
		log.Fatalf("Got error connecting: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewChatClient(conn)

	sign(c)

	communicate(c)
}

func sign(c pb.ChatClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	fmt.Println("# You need to sign first. Please enter your name below")
	b := bufio.NewReader(os.Stdin)
	in, err := b.ReadBytes('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}
	user := strings.Trim(string(in), "\r\n")
	res, err := c.Sign(ctx, &pb.SignRequest{Name: user})
	if err != nil {
		log.Fatalf("Signing error: %v\n", err)
	}
	fmt.Printf("# You're signed, %v. Proceeding to chat...\n", user)
	fmt.Println("# Type 'help' to get the list of commands")
	uid = res.UserId
}

func communicate(c pb.ChatClient) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wmeta := metadata.NewOutgoingContext(ctx, metadata.Pairs("uid", uid))
	stream, err := c.Communicate(wmeta)
	if err != nil {
		log.Fatalf("Error connecting to Communicate rpc: %v\n", err)
	}

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				// end of communication
				return
			}
			if err != nil {
				log.Fatalf("Error receiving: %v\n", err)
			}
			fmt.Printf("[%v] %v: %v\n", res.SentAt.AsTime().Format(time.RFC822), res.User, res.Message)
		}
	}()

Receive:
	for {
		b := bufio.NewReader(os.Stdin)
		in, err := b.ReadBytes('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v\n", err)
		}
		phrase := strings.Trim(string(in), "\r\n")

		switch phrase {
		case "help":
			fmt.Println("# List of commands")
			fmt.Println("# q: quit")
		case "q":
			fmt.Println("# good bye")
			break Receive
		default:
			stream.Send(&pb.Message{Message: phrase})
		}
	}

	stream.CloseSend()
}
