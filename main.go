package main

import (
	"fmt"
	"log"
	"net"
	"strings"

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
	reg := command.NewRegistry()        // Command router
	command.RegisterBasic(reg)          // Register basic commands, PING, SET, GET, DELETE

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

	r := resp.NewResp(conn)             // decoder
	w := resp.NewWriter(conn)           // encoder
	ctx := &command.Ctx{Storage: store} // Create a new context with the storage

	for {
		req, err := r.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received request:", req)

		cmd, args, err := toCommand(req)
		if err != nil {
			fmt.Println("Failed to convert request to execute command")
			return
		}
		reply := reg.Exec(ctx, cmd, args)

		if err := w.Write(reply); err != nil {
			fmt.Println("Error writing response:", err)
			return
		}
	}
}

func toCommand(v resp.Value) (string, []resp.Value, error) {
	if v.Typ != "array" || len(v.Array) == 0 {
		return "", nil, fmt.Errorf("ERR: expected array")
	}

	head := v.Array[0]
	var cmdName string

	switch head.Typ {
	case "bulk":
		cmdName = head.Bulk
	case "string":
		cmdName = head.Str
	}

	args := v.Array[1:]
	return strings.ToUpper(cmdName), args, nil
}
