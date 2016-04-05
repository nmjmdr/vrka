package main

// sample client for vrka

import (
	"fmt"
	"os"
	"net"
)


func main() {

	args := os.Args

	if args == nil || len(args) < 2 {
		fmt.Println("Usage: client <server-address>")
		return
	}

	addr := args[1]

	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)

	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		fmt.Println("Connection to server failed:", err.Error())
		return
	}

	addCmd := "+\r\n10000\r\nhttps://google.com"
	_, err = conn.Write([]byte(addCmd))
	if err != nil {
		fmt.Println("Send  to server failed:", err.Error())
		return
        }

	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		println("Read response from server failed:", err.Error())
		return
	}
	fmt.Println("Reply from server=", string(reply))

}
