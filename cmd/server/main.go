package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishu-commits7/distributed-raft-storage/pkg"
)

func main() {
	var (
		nodeID    = flag.String("id", "node1", "Node ID")
		dataDir   = flag.String("data", "/tmp/raft-node", "Data directory")
		bindAddr  = flag.String("addr", "127.0.0.1", "Bind address")
		bindPort  = flag.Int("port", 5000, "Bind port")
		grpcPort  = flag.Int("grpc-port", 6000, "gRPC port")
		bootstrap = flag.Bool("bootstrap", false, "Bootstrap cluster")
	)
	flag.Parse()

	// Create Raft configuration
	ratCfg := &pkg.RaftConfig{
		NodeID:    *nodeID,
		DataDir:   *dataDir,
		BindAddr:  *bindAddr,
		BindPort:  *bindPort,
		Bootstrap: *bootstrap,
	}

	// Create Raft engine
	engine, err := pkg.NewRaftEngine(ratCfg)
	if err != nil {
		log.Fatalf("Failed to create Raft engine: %v", err)
	}
	defer engine.Close()

	log.Printf("Raft node started: %s (Leader: %v)", *nodeID, engine.IsLeader())

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *bindAddr, *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	server := pkg.NewGRPCServer(engine)
	go func() {
		log.Printf("gRPC server listening on %s:%d", *bindAddr, *grpcPort)
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	server.GracefulStop()
}
