package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/tinyQQ/util"
)

var (
	loginName string
)

func main() {
	//fmt.Printf("请输入用户名：")
	//fmt.Scan(&loginName)
	conn, err := net.Dial("tcp", ":8888")
	defer conn.Close()

	if err != nil {
		log.Printf("dial error: %s", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	closeSignal := make(chan struct{})
	//read msg
	go readMessage(conn, &wg, closeSignal)

	//_, err = conn.Write([]byte(loginName))
	//_, err = util.Write(conn, loginName)
	if err != nil {
		log.Printf("login error: %s", err)
		return
	}
	//receive and send msg
	go receiveContentFromStdin(conn, closeSignal)
	wg.Wait()
}

func readMessage(conn net.Conn, wg *sync.WaitGroup, closeSignal chan struct{}) {
	defer wg.Done()
	for {
		//buf := make([]byte, 1024)
		var readContent string
		//fmt.Println("read from host")
		readContent, err := util.Read(conn)

		if err != nil {
			log.Printf("read message error: %s", err)
			//host close the connection
			if err == io.EOF {
				closeSignal <- struct{}{}
				fmt.Println("对方已经下线，聊天中断~")
				break
			}
			continue
		}

		fmt.Printf(readContent)
		fmt.Println()
		//if toName == "" {
		//	fmt.Println("请输入对方用户名：")
		//}
	}
}

func receiveContentFromStdin(conn net.Conn, closeSignal chan struct{}) {
//loop:
//	for {
//		select {
//		case <-closeSignal:
//			break loop
//		default:
//
//		}
//		buf := make([]byte, 1024)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			content := scanner.Text()
			util.Write(conn, content)
		}
		if err := scanner.Err(); err != nil {
			log.Println("Receive from stdin error:", err)
		}
}
