package main

import (
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/joho/godotenv/autoload"
	pb "github.com/slayerbirden/chat/gen"
	"github.com/slayerbirden/chat/pkg/interceptors"
	"github.com/slayerbirden/chat/pkg/repo"
	"github.com/slayerbirden/chat/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// create chat server
	r := repo.NewUsersMap(
		make(map[string]repo.User),
		make(map[string]string),
	)
	cs := server.NewChatServer(r)
	s := grpc.NewServer(
		grpc.StreamInterceptor(interceptors.UsersStreamServerInterceptor(r)),
	)
	pb.RegisterChatServer(s, cs)
	reflection.Register(s)
	defer func() {
		cs.Exit()
		s.GracefulStop()
	}()

	log.Fatal(serve(s))
}

func serve(s *grpc.Server) error {
	os.Setenv("URI", "localhost:9090")
	li, err := net.Listen("tcp", os.Getenv("URI"))
	if err != nil {
		return err
	}
	defer li.Close()

	fmt.Println("Listening on", os.Getenv("URI"), "...")
	if err := s.Serve(li); err != nil {
		return err
	}
	return nil
}
