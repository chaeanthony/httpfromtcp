package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	port = ":42069"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatalf("error resolving UDP: %s", err.Error())
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("error dialing UDP: %s", err.Error())
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("error reading input: %s", err.Error())
		}

		_, err = udpConn.Write([]byte(text))
		if err != nil {
			log.Fatalf("error writing to UDP: %s", err.Error())
		}
	}
}
