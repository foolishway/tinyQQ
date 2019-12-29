package main

import (
	// "bytes"

	"fmt"
	"io"

	//"time"

	"github.com/tinyQQ/util"
	//"io"
	"log"
	"net"
	"strings"
)

type LoginUser struct {
	Conn        net.Conn
	chatter string
	Addr string
}

var userMap = make(map[string][]LoginUser)

const (
	remindMsg string = "请输入聊天对象："
)
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
	var (
		lUser, addr string
	)
	for {
		//fmt.Printf("%s read from %q", lUser, conn)
		// read from the connection
		//n, err := conn.Read(buf)
		//conn.SetReadDeadline(time.Now().Add(time.Microsecond * 50))
		readContent, err := util.Read(conn)
		//readContent = readContent[:len(readContent) - 1]
		//fmt.Printf("read:%s\n", readContent)
		//log.Printf("err:%s\n", err)
		if err != nil {
			//log.Printf("conn read error: %s", err)
			//if err, ok := err.(net.Error); ok && err.Timeout() {
			//	continue
			//}
			////client closed
			if err == io.EOF {
				fmt.Printf("%s(%s)已下线~", lUser, addr)
				addr := conn.RemoteAddr().String()
				offLine(lUser, addr)
				return
			}
		}

		//login, record the user
		if lUser == "" {
			fmt.Printf("%s已上线~\n", readContent)
			lUser = readContent
			if _, ok := userMap[string(lUser)]; !ok {
				//user address eg:127.0.0.1:8000
				addr = conn.RemoteAddr().String()
				userMap[lUser] = []LoginUser{LoginUser{Conn: conn, Addr: addr}}
				sendLoginUsersToLUser(lUser, conn)
				continue
			} else {
				for i := 0; i < len(userMap[lUser]); i++ {
					if userMap[lUser][i].chatter == "" {
						util.Write(conn,"该用户存在已打开的且未选择聊天对象的终端~")
						return
					}
				}
			}
		}

		transportMessage(lUser, addr, readContent, conn)
	}
}

func transportMessage(lUser, addr, content string, conn net.Conn) {
	u := userMap[lUser]
	if _, ok := userMap[lUser]; ok {
		for i := 0; i < len(u); i++ {
			if u[i].Addr == addr {
				//unbind chatter
				if u[i].chatter == "" {
					if lUser == content {
						msg := fmt.Sprintf("系统消息：不能和自己聊天~\n%s", remindMsg)
						util.Write(conn, msg)
						break
					} else {
						//bind chatter
						u[i].chatter = content
						msg := fmt.Sprintf("系统消息：连接已建立，和%s快乐的聊天吧~", content)
						util.Write(conn, msg)
						break
					}
				} else {
					//transport message
					chatter := u[i].chatter
					chatterConn := getConnByUser(chatter, lUser)
					util.Write(chatterConn, fmt.Sprintf("%s：%s",lUser, content))
					break
				}
			}
		}
	}
}

func getConnByUser(tUser, lUser string) (conn net.Conn) {
	if _, ok := userMap[tUser]; ok {
		for i := 0; i < len(userMap[tUser]); i++ {
			if userMap[tUser][i].chatter == lUser || userMap[tUser][i].chatter == "" {
				//fmt.Println("find conn~")
				conn = userMap[tUser][i].Conn
			}
		}
	}
	return
}

func sendLoginUsersToLUser(lUser string, conn net.Conn) {
	var loginUsers []string
	for k, _ := range userMap {
		if lUser != k {
			loginUsers = append(loginUsers, k)
		}
	}
	var msg string = "暂无好友~"
	if len(loginUsers) > 1 {
		msg = fmt.Sprintf("当前在线好友：%s%c%s", strings.Join(loginUsers, ","), '\n', remindMsg)
	}
	if len(loginUsers) == 1 {
		msg = fmt.Sprintf("当前在线好友：%s%c%s", loginUsers[0],'\n', remindMsg)
	}
	util.Write(conn, msg)
}

//off line
func offLine(lUser, addr string) {
	reportTUserOffLine(lUser, addr)
	closeConn(lUser, addr)
}

func closeConn(lUser, addr string) {
		fmt.Println("closeConn~")
		if _, ok := userMap[lUser]; ok {
			if len(userMap[lUser]) <= 1 {
				//close the connection
				userMap[lUser][0].Conn.Close()
				delete(userMap, lUser)
				fmt.Println("closeConn2~")

				fmt.Printf("userMap:%q", userMap)
				return
			}


			for i := 0; i < len(userMap[lUser]); i++ {
				if userMap[lUser][i].Addr == addr {
					// delete(userMap[lUser][i])
					userMap[lUser][i].Conn.Close()
					userMap[lUser] = append(userMap[lUser][:i], userMap[lUser][i+1:]...)
					break
				}
			}
		}
		fmt.Printf("userMap:%q", userMap)

}

func reportTUserOffLine(lUser, addr string)  {
	fmt.Println("reportTUserOffLine", addr)
	var chatter string
	var conn net.Conn
	if _, ok := userMap[lUser]; ok {
		for i := 0; i < len(userMap[lUser]); i++ {
			if userMap[lUser][i].Addr == addr {
				chatter = userMap[lUser][i].chatter
				break
			}
		}
		//for i := 0; i < len(userMap[tUser]); i++ {
		//	if userMap[tUser][i].chatter == lUser {
		//		userMap[tUser][i].chatter = ""
		//		msg := fmt.Sprintf("%s已下线\n%s", lUser, remindMsg)
		//
		//		util.Write(userMap[tUser][i].Conn, msg)
		//	}
		//}
	}

	fmt.Println("chatter:", chatter)
	if chatter != "" {
		if _, ok := userMap[chatter]; ok {
			for i := 0; i < len(userMap[chatter]); i++ {
				if userMap[chatter][i].chatter == lUser {
					conn = userMap[chatter][i].Conn
					break
				}
			}
		}
	}

	if conn != nil {
		msg := fmt.Sprintf("%s已下线\n%s", lUser, remindMsg)

		util.Write(conn, msg)
	}
}