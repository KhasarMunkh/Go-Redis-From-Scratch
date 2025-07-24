package main

import (
	"fmt"
	"log"
	"net"

	"github.com/KhasarMunkh/Go-Redis-From-Scratch/resp" // Importing the resp package for RESP protocol handling
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

	for {
		// listen for connections

		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	r := resp.NewResp(conn)        // decoder
	w := resp.NewWriter(conn) // encoder
	for {
		val, err := r.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(val)
		w.Write(resp.Value{Typ: "string", Str: "Hello!"})
	}
}
