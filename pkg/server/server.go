package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	pb "github.com/slayerbirden/chat/gen"
	"github.com/slayerbirden/chat/pkg/repo"
	"github.com/slayerbirden/chat/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	internalError = "There was an error internally"
)

var channels map[string]chan *pb.Envelope = make(map[string]chan *pb.Envelope)

// GetChannels returns current channels
func GetChannels() map[string]chan *pb.Envelope {
	return channels
}

// AddChannel adds a channel
func AddChannel(id string, c chan *pb.Envelope) {
	channels[id] = c
}

// ChatServer adds users to chat and facilitates the communication
type ChatServer struct {
	pb.UnimplementedChatServer

	r repo.Users
}

// NewChatServer constructor
func NewChatServer(r repo.Users) *ChatServer {
	return &ChatServer{
		r: r,
	}
}

// Sign "registers" user and allows to communicate in the chat
func (s *ChatServer) Sign(ctx context.Context, r *pb.SignRequest) (*pb.SignResponse, error) {
	u, err := s.r.AddUser(r.Name)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.SignResponse{
		UserId: u.ID,
	}, nil
}

// Communicate deals with the communication between clients
func (s *ChatServer) Communicate(stream pb.Chat_CommunicateServer) error {
	chanID, err := util.GetSingleMetaFromStream(stream, "uid")
	if err != nil {
		return err
	}
	u, err := s.r.GetUser(chanID)
	if err != nil {
		log.Printf("Faild to find the user in repo. ID: %v\n", chanID)
		return status.Error(codes.Internal, internalError)
	}
	messages := channels[chanID]
	e := make(chan error)
	defer close(e)
	go func() {
		select {
		case <-stream.Context().Done():
			log.Println("client exited")

			for k, c := range channels {
				// only send to other receivers
				if k != chanID {
					c <- &pb.Envelope{
						Id:      uuid.New().String(),
						Message: fmt.Sprintf("User %v logged out", u.Name),
						User:    "system",
						SentAt:  timestamppb.Now(),
					}
				}
			}
		}
	}()
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				delete(channels, chanID)
				close(messages)
				fmt.Println("Client stopped streaming")
				e <- nil
				return
			}
			if err != nil {
				log.Printf("Got error receiving message: %v\n", err)
				delete(channels, chanID)
				close(messages)
				e <- errors.New("Error while receiving")
				return
			}
			// send Envelope to all channels
			for k, c := range channels {
				// only send to other receivers
				if k != chanID {
					c <- &pb.Envelope{
						Message: msg.GetMessage(),
						User:    u.Name,
						SentAt:  timestamppb.Now(),
						Id:      uuid.New().String(),
					}
				}
			}
		}
	}()
	go func() {
		for m := range messages {
			err := stream.Send(m)
			if err != nil {
				log.Printf("Got error sending envelope: %v", err)
				e <- errors.New("Error while sending")
				return
			}
		}
	}()
	return <-e
}

// Who sends a stream of users
func (s *ChatServer) Who(r *emptypb.Empty, stream pb.Chat_WhoServer) error {
	for _, u := range s.r.ListUsers() {
		err := stream.Send(&pb.WhoResponse{Name: u.Name})
		if err != nil {
			log.Printf("Error while getting users: %v", err)
			return status.Error(codes.Internal, internalError)
		}
	}
	return nil
}

// Exit does some graceful exiting
func (s *ChatServer) Exit() {
	// close all channels
	for _, c := range channels {
		close(c)
	}
}
