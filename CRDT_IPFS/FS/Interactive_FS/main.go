package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	id := flag.String("id", "r1", "Replica ID")
	port := flag.String("port", "8001", "Port to listen on")
	peerPort := flag.String("peer", "8002", "Peer's port")
	flag.Parse()

	replica := &Replica{
		ID:      *id,
		NodeSet: NewNodeSetCRDT(),
	}

	replica.AddNode("/", "root", "", Dir)

	// Start server to receive merges
	go startServer(*port, replica)

	// Start periodic merge with peer
	go func() {
		for {
			time.Sleep(60 * time.Second)
			mergeWithPeer(*peerPort, replica)
		}
	}()

	// CLI input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter command (add <path> <type> | remove <path> <type> | print): ")
		scanner.Scan()
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "add":
			if len(parts) < 3 {
				fmt.Println("Usage: add <path> <dir|file>")
				continue
			}
			pathStr := parts[1]
			typeStr := parts[2]
			var nodeType NodeType
			if strings.EqualFold(typeStr, "dir") {
				nodeType = Dir
			} else {
				nodeType = File
			}
			parent := path.Dir(pathStr)
			name := path.Base(pathStr)
			// ts := replica.NextTimestamp()
			// node := NewNode(path, name, parent, nodeType, ts, replica.ID)
			fmt.Println(pathStr, name, parent, nodeType)
			replica.AddNode(pathStr, name, parent, nodeType)
		case "remove":
			if len(parts) < 3 {
				fmt.Println("Usage: remove <path> <dir|file>")
				continue
			}
			pathStr := parts[1]
			typeStr := parts[2]
			var nodeType NodeType
			if typeStr == "dir" {
				nodeType = Dir
			} else {
				nodeType = File
			}
			// ts := replica.NextTimestamp()
			replica.RemoveNode(pathStr, nodeType)
		case "print":
			tree := BuildTree(replica.NodeSet, "root")
			// ResolveNameConflicts(tree, replica.NodeSet)
			PrintTree(tree, "")
		case "final":
			var policy string
			if len(parts) == 2 {
				policy = parts[1]
				fmt.Printf("Orphan node policy: %s\n", policy)
			} else {
				fmt.Println("Default orphan node policy: skip")
				policy = "skip"
			}
			fmt.Println("\n------------------FINAL TREE---------------")
			mergeWithPeer(*peerPort, replica)
			time.Sleep(500 * time.Millisecond)
			tree := FinalTree(replica.NodeSet, policy)
			PrintTree(tree, "")
			os.Exit(0)
		default:
			fmt.Println("Unknown command")
		}
	}
}

func startServer(port string, replica *Replica) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, replica)
	}
}

func handleConnection(conn net.Conn, replica *Replica) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	var incoming NodeSetCRDT
	if err := dec.Decode(&incoming); err != nil {
		log.Println("Failed to decode:", err)
		return
	}
	replica.NodeSet.Merge(&incoming)
	fmt.Println("[MERGE] Merged with peer")
}

func mergeWithPeer(peerPort string, replica *Replica) {
	conn, err := net.Dial("tcp", ":"+peerPort)
	if err != nil {
		return
	}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	_ = enc.Encode(replica.NodeSet)
}
