package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/vishu-commits7/distributed-raft-storage/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [options]")
		fmt.Println("Commands: get, set, delete, scan, join, leave, status")
		os.Exit(1)
	}

	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)

	addr := fs.String("addr", "127.0.0.1:6000", "Server address")

	switch cmd {
	case "get":
		key := fs.String("key", "", "Key to get")
		fs.Parse(os.Args[2:])
		if *key == "" {
			log.Fatal("Key is required")
		}
		cmdGet(*addr, *key)

	case "set":
		key := fs.String("key", "", "Key to set")
		value := fs.String("value", "", "Value to set")
		fs.Parse(os.Args[2:])
		if *key == "" || *value == "" {
			log.Fatal("Key and value are required")
		}
		cmdSet(*addr, *key, *value)

	case "delete":
		key := fs.String("key", "", "Key to delete")
		fs.Parse(os.Args[2:])
		if *key == "" {
			log.Fatal("Key is required")
		}
		cmdDelete(*addr, *key)

	case "scan":
		prefix := fs.String("prefix", "", "Prefix to scan")
		limit := fs.Int("limit", 0, "Limit results")
		fs.Parse(os.Args[2:])
		cmdScan(*addr, *prefix, int32(*limit))

	case "join":
		nodeID := fs.String("id", "", "Node ID to join")
		nodeAddr := fs.String("node-addr", "", "Node address to join")
		fs.Parse(os.Args[2:])
		if *nodeID == "" || *nodeAddr == "" {
			log.Fatal("Node ID and address are required")
		}
		cmdJoin(*addr, *nodeID, *nodeAddr)

	case "status":
		fs.Parse(os.Args[2:])
		cmdStatus(*addr)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func getClient(addr string) (pb.StorageServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewStorageServiceClient(conn)
	return client, conn, nil
}

func cmdGet(addr, key string) {
	client, conn, err := getClient(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	resp, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}

	if resp.Found {
		fmt.Printf("%s\n", resp.Value)
	} else {
		fmt.Println("Key not found")
	}
}

func cmdSet(addr, key, value string) {
	client, conn, err := getClient(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	resp, err := client.Set(context.Background(), &pb.SetRequest{Key: key, Value: value})
	if err != nil {
		log.Fatalf("Failed to set: %v", err)
	}

	if resp.Success {
		fmt.Println("OK")
	} else {
		fmt.Printf("Error: %s\n", resp.Error)
	}
}

func cmdDelete(addr, key string) {
	client, conn, err := getClient(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	resp, err := client.Delete(context.Background(), &pb.DeleteRequest{Key: key})
	if err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}

	if resp.Success {
		fmt.Println("OK")
	} else {
		fmt.Printf("Error: %s\n", resp.Error)
	}
}

func cmdScan(addr, prefix string, limit int32) {
	client, conn, err := getClient(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	stream, err := client.Scan(context.Background(), &pb.ScanRequest{Prefix: prefix, Limit: limit})
	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Printf("%s\t%s\n", resp.Key, resp.Value)
	}
}

func cmdJoin(addr, nodeID, nodeAddr string) {
	// This would typically call an admin RPC to join a peer
	fmt.Printf("Join command not yet implemented. Would add %s at %s to cluster via %s\n", nodeID, nodeAddr, addr)
}

func cmdStatus(addr string) {
	client, conn, err := getClient(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Attempt a ping via Get
	resp, err := client.Get(context.Background(), &pb.GetRequest{Key: "__status__"})
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}

	fmt.Printf("Server is online. Response: %v\n", resp)
}
