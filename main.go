package main

import (
	"vrka"
	"net"
	"fmt"
	"os"
	"strings"
	"errors"
	"strconv"
)


func main() {

	args := os.Args

	port := "9091"

	if args != nil && len(args) >= 2 {
		port = args[2]
	}

	portString := ":"+port

	fmt.Printf("starting vrka on %s",port)
	ln, _ := net.Listen("tcp", portString)  
	
	v := vrka.New()
 
	for {		
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn,v)
	}	
}

func handleRequest(conn net.Conn,v vrka.Vrka) {

	buf := make([]byte, 1024)  
	n, err := conn.Read(buf)
	
	if err != nil || n == 0 {
		conn.Close()
		return
	}
	smsg := string(buf[0:n])

	sp := strings.Fields(smsg)
	if sp == nil || len(sp) < 2 {
		err := "bad request"
		conn.Write([]byte(err))
		return
	}

	response := ""	
	err = nil
	if sp[0] == "+" && len(sp) == 3 {
		response,err = handleAdd(sp[1],sp[2],v)		
	} else if sp[1] == "-" && len(sp) == 2 {
		err = handleDel(sp[1],v)		
	} else {
		err = errors.New("Bad request")
	}

	reply := ""
	if err != nil {
		reply = "!\r\n" + err.Error()
	} else {
		reply = "*\r\n" + response
	}

	conn.Write([]byte(reply))	 
	conn.Close()	
}
// not considering the payload for now
// assuming that the server would do 
// a GET request
func handleAdd(afterms string, uri string, v vrka.Vrka) (string,error) {
	
	u, err := strconv.ParseUint(afterms, 10, 64)
	if err != nil {
		return "", err
	}
	id,err := v.Add(uri,"",u)
	return id, err
}

func handleDel(id string, v vrka.Vrka) error {
	err := v.Del(id)
	return err
}
