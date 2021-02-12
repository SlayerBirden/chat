package util

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GetSingleMetaFromStream simplifies retrieving things from stream metadata
func GetSingleMetaFromStream(s grpc.ServerStream, key string) (string, error) {
	m, ok := metadata.FromIncomingContext(s.Context())
	if !ok {
		return "", status.Error(codes.Internal, "Could not find Metadata from context")
	}
	res, ok := m[key]
	if !ok {
		return "", status.Error(codes.PermissionDenied, fmt.Sprintf("Missing %v in meta", key))
	}
	if len(res) != 1 {
		return "", status.Error(codes.InvalidArgument, fmt.Sprintf("Metadata value is not single for %v", key))
	}
	return res[0], nil
}
