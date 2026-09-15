package pkg

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/vishu-commits7/distributed-raft-storage/pkg/proto"
	"google.golang.org/grpc"
)

// GRPCServer implements the StorageService
type GRPCServer struct {
	engine *RaftEngine
	pb.UnimplementedStorageServiceServer
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(engine *RaftEngine) *grpc.Server {
	grpcServer := grpc.NewServer()
	pb.RegisterStorageServiceServer(grpcServer, &GRPCServer{engine: engine})
	return grpcServer
}

// Get retrieves a value from the storage
func (s *GRPCServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, found := s.engine.Get(req.Key)
	return &pb.GetResponse{
		Value: val,
		Found: found,
	}, nil
}

// Set sets a value in the storage
func (s *GRPCServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if err := s.engine.Set(req.Key, req.Value); err != nil {
		return &pb.SetResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SetResponse{
		Success: true,
	}, nil
}

// Delete deletes a key from the storage
func (s *GRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := s.engine.Delete(req.Key); err != nil {
		return &pb.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.DeleteResponse{
		Success: true,
	}, nil
}

// Scan returns all keys with the given prefix
func (s *GRPCServer) Scan(req *pb.ScanRequest, stream grpc.ServerStream) error {
	results := s.engine.Scan(req.Prefix)

	count := 0
	for k, v := range results {
		if req.Limit > 0 && count >= int(req.Limit) {
			break
		}

		if err := stream.SendMsg(&pb.ScanResponse{
			Key:   k,
			Value: v,
		}); err != nil {
			return err
		}
		count++
	}

	return nil
}

// Health check endpoint
func (s *GRPCServer) Health(w http.ResponseWriter, r *http.Request) {
	if s.engine.IsLeader() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Leader: %s", s.engine.GetLeader())))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Follower (Leader: %s)", s.engine.GetLeader())))
	}
}
