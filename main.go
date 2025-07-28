package main

import (
	"fmt"
	"log"
	"net"

	"github.com/KhasarMunkh/Go-Redis-From-Scratch/command" // Importing the command package for command handling
	"github.com/KhasarMunkh/Go-Redis-From-Scratch/resp"    // Importing the resp package for RESP protocol handling
	"github.com/KhasarMunkh/Go-Redis-From-Scratch/storage" // Importing the storage package for data storage
)

func main() {
	fmt.Println("Listenting on port: 6379")

	// create listener object
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	// Create shared data state for server
	store := storage.NewMemoryStorage() // Create a new in-memory storage, shared keyspace
	reg := command.NewRegistry()          // Command router
	command.RegisterBasic(reg)            // Register basic commands, PING, SET, GET, DELETE

	for {
		// listen for connections

		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		handleConnection(conn, reg, store)

	}
}

func handleConnection(conn net.Conn, reg *command.Registry, store storage.Storage) {
	defer conn.Close()

	r := resp.NewResp(conn)           // decoder
	w := resp.NewWriter(conn)         // encoder
	ctx := &command.Ctx{Storage: store} // Create a new context with the storage

	for {
		req, err := r.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received request:", req)

		cmd, args := toCommand(req)
		reg.Exec(ctx, cmd, args)

		w.Write(resp.Value{Typ: "string", Str: "Hello!"})
	}
}

func toCommand(req resp.Value) (string, []resp.Value) {
	if req.Typ != "array" || len(req.Array) == 0 {
		return "", nil
	}
	cmd := req.Array[0].Str
	args := req.Array[1:]
	return cmd, args
}

