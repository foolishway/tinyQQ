package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type LoginUser struct {
	Conn        net.Conn
	AnotherSide string
}

var userMap map[string][]LoginUser

func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Println("error listen:", err)
		return
	}
	defer l.Close()
	log.Println("listen ok")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			break
		}
		//one goroutine per user
		go handleConn(conn)
	}
}
func handleConn(conn net.Conn) {
	defer conn.Close()
loop:
	for {
		// read from the connection
		var buf = make([]byte, 200)
		var (
			lUser, tUser, data []byte
		)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("conn read %d bytes,  error: %s", n, err)
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			//other side closed
			if err == io.EOF {

				break
			}
		}

		lu, tu, data, err := unpacket(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		//login, record the user
		if len(lUser) == 0 {
			lUser = lu
			us := string(lUser)
			if _, ok := userMap[us]; !ok {
				userMap[us] = []LoginUser{LoginUser{Conn: conn}}
				sendLoginUsersToLUser(string(lUser), conn)
				continue
			} else {
				for i := 0; i < len(userMap[us]); i++ {
					if userMap[us][i].AnotherSide == "" {
						conn.Write([]byte("该用户存在已打开的且未选择聊天对象的终端~"))
						break loop
					}
				}
			}
		}
		//select chat
		if len(tUser) == 0 && len(tu) != 0 {
			tUser = tu
			us := string(lUser)
			tus := string(tUser)
			if us == tus {
				conn.Write([]byte("不能和自己聊天~"))
				break loop
			}
			if _, ok := userMap[us]; ok {
				for i := 0; i < len(userMap[us]); i++ {
					if userMap[us][i].AnotherSide == "" {
						userMap[us][i].AnotherSide = tus
						conn.Write([]byte(fmt.Sprintf("开始和%s畅聊吧~", tus)))
						break
					}
				}
			}
			continue
		}

		//send data to tUsers
		if len(data) > 0 {
			tConn := getConnByUser(string(tUser), string(lUser))
			if tConn != nil {
				tConn.Write(data)
			}
		}
		// log.Printf("read %d bytes, content is %s\n", n, string(buf[:n]))
	}
}

func unpacket(packetData []byte) (lUser, tUser, data []byte, err error) {

	incorrectError := errors.New("Unpacket error: incorrect format packet data.")

	if len(packetData) < 50 {
		// err = incorrectError
		data = packetData
		return
	}

	packetHeader := packetData[:50]
	data = packetData[50:]

	if bytes.Count(packetHeader, []byte{'|'}) != 2 {
		err = incorrectError
		return
	}

	bs := bytes.Split(packetHeader, []byte{'|'})
	lUser, tUser = bs[0], bs[1]
	return
}

func getConnByUser(tUser, lUser string) (conn net.Conn) {
	if _, ok := userMap[tUser]; ok {
		for i := 0; i < len(userMap[tUser]); i++ {
			if userMap[tUser][i].AnotherSide == lUser {
				conn = userMap[tUser][i].Conn
			}
		}
	}
	return
}

func sendLoginUsersToLUser(lUser string, conn net.Conn) {
	var loginUsers []string
	for k, _ := range userMap {
		loginUsers = append(loginUsers, k)
	}
	conn.Write([]byte(strings.Join(loginUsers, ",")))
}
