package main

import (
	"vrka"
	"net"
	"fmt"
	"bufio"
	"os"
	"strings"
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
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn,v)
	}	
}

// Handles incoming requests.
func handleRequest(conn net.Conn,v vrka.Vrka) {
	msg, _ := bufio.NewReader(conn).ReadString('\r\n')     
	smsg := string(msg)  

	sp := strings.Fields(smsg)
	if sp == nil || len(sp) < 2 {
		err := "bad request"
		conn.Write([]byte(err))
		return
	}

	response := ""
	var err error
	err = nil
	if sp[0] == "+" && len(sp) == 4 {
		response,err = handleAdd(sp[1],sp[2],sp[3],v)		
	} else if sp[1] == "-" && len(sp) == 2 {
		err = handleDel(sp[1],sp[2].v)		
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


func handleAdd(afterms string, uri string,payload string, v vrka.Vrka) (string,error) {
	
	u, err := strconv.ParseUint(afterms, 10, 64)
	if err != nil {
		return "", err
	}
	id,err := v.Add(uri,payload,u)
	return id, err
}

func handleDel(id string, v vrka.Vrka) error {
	err := v.Del(id)
	return err
}
