package interceptors

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	pb "github.com/slayerbirden/chat/gen"
	"github.com/slayerbirden/chat/pkg/repo"
	"github.com/slayerbirden/chat/pkg/server"
	"github.com/slayerbirden/chat/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UsersStreamServerInterceptor "registers users for the streaming communication using metadata"
func UsersStreamServerInterceptor(r repo.Users) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// only check user for Communicate
		method := info.FullMethod
		if strings.HasSuffix(method, "Communicate") {
			uID, err := util.GetSingleMetaFromStream(ss, "uid")
			if err != nil {
				return err
			}
			u, err := r.GetUser(uID)
			if err != nil {
				return status.Error(codes.InvalidArgument, "User is not signed")
			}
			messages := make(chan *pb.Envelope)
			server.AddChannel(uID, messages)
			// notify
			channels := server.GetChannels()
			for i, c := range channels {
				if i != uID {
					c <- &pb.Envelope{
						Id:      uuid.New().String(),
						Message: fmt.Sprintf("user %v entered the chat", u.Name),
						User:    "system",
						SentAt:  timestamppb.Now(),
					}
				}
			}
		}

		return handler(srv, ss)
	}
}
