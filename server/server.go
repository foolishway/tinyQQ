package main

import (
	// "bytes"

	"fmt"
	"io"

	//"time"

	"github.com/tinyQQ/util"
	//"io"
	"errors"
	"log"
	"net"
	"strings"
)

type loginUser struct {
	Addr    string
	Conn    net.Conn
	chatter string
	ChatterAddr string
}

var userMap = make(map[string][]loginUser)

const (
	remindMsg1 string = "请输入用户名："
	remindMsg2 string = "请输入聊天对象："
	remindMsg3 string = "系统消息：%s找您聊天，您已经和TA建立了连接，可以畅聊了~"
	remindMsg4 string = "未找到您输入的聊天对象，请重新输入："
	remindMsg5 string = "系统消息：连接已建立，和%s快乐的聊天吧~"
	remindMsg6 string = "系统消息：不能和自己聊天~\n%s"
	remindMsg7 string = "该用户存在已打开的且未选择聊天对象的终端~\n%s"

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
	if len(lUser) == 0 {
		util.Write(conn, remindMsg1)
	}
loop:
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
				fmt.Printf("%s(%s)已下线~\n", lUser, addr)
				addr := conn.RemoteAddr().String()
				offLine(lUser, addr)
				return
			}
		}

		//login, record the user
		if lUser == "" {
			fmt.Printf("%s已上线~\n", readContent)
			addr = conn.RemoteAddr().String()
			//if _, ok := userMap[readContent]; !ok {
			//	//user address eg:127.0.0.1:8000
			//	lUser = readContent
			//	userMap[lUser] = []loginUser{loginUser{Conn: conn, Addr: addr}}
			//	sendLoginUsersToLUser(lUser, conn)
			//	continue
			//} else {
				//indicate whether login user is waiting for the chatter
				var isWaiting bool
				for i := 0; i < len(userMap[readContent]); i++ {
					if userMap[readContent][i].chatter == "" {
						util.Write(conn, fmt.Sprintf(remindMsg7, remindMsg1))
						isWaiting = true
						continue loop
					}
				}
				if !isWaiting {
					lUser = readContent
					userMap[lUser] = append(userMap[lUser], loginUser{Conn: conn, Addr: addr})
					sendLoginUsersToLUser(lUser, conn)
					continue
				}
			//}
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
						msg := fmt.Sprintf(remindMsg6, remindMsg2)
						util.Write(conn, msg)
						return
					} else {
						tUser := content
						//bind chatter
						chatterAddr, err := forceConnection(lUser, tUser, addr)
						if err != nil {
							util.Write(conn, err.Error())
							return
						}
						//chatterAddr := getAddrByUser(tUser)
						u[i].chatter = content
						u[i].ChatterAddr = chatterAddr
						msg := fmt.Sprintf(remindMsg5, content)
						util.Write(conn, msg)
						return
					}
				} else {
					//transport message
					chatter := u[i].chatter
					chatterAddr := u[i].ChatterAddr
					chatterConn := getConnByUser(chatter, chatterAddr)
					util.Write(chatterConn, fmt.Sprintf("%s：%s", lUser, content))
					return
				}
			}
		}
	}
}
func getConnByUser(chatter, chatterAddr string) (conn net.Conn) {
	if _, ok := userMap[chatter]; ok {
		for i := 0; i < len(userMap[chatter]); i++ {
			//if (userMap[tUser][i].chatter == lUser && userMap[tUser][i].ChatterAddr == chatterAddr)  || userMap[tUser][i].chatter == "" {
			if userMap[chatter][i].Addr == chatterAddr {
				//fmt.Println("find conn~")
				conn = userMap[chatter][i].Conn
			}
		}
	}
	return
}

//tell user that who is waiting to connect now
func sendLoginUsersToLUser(lUser string, conn net.Conn) {
	var loginUsers []string
	for k, us := range userMap {
		var isWaiting bool
		for _, v := range us {
			if v.chatter == "" && lUser != k {
				isWaiting = true
				break
			}
		}
		if isWaiting {
			loginUsers = append(loginUsers, k)
		}
	}
	var msg string = "暂无好友~"
	if len(loginUsers) > 1 {
		msg = fmt.Sprintf("当前在线好友：%s%c%s", strings.Join(loginUsers, ","), '\n', remindMsg2)
	}
	if len(loginUsers) == 1 {
		msg = fmt.Sprintf("当前在线好友：%s%c%s", loginUsers[0], '\n', remindMsg2)
	}
	util.Write(conn, msg)
}

//off line
func offLine(lUser, addr string) {
	reportTUserOffLine(lUser, addr)
	closeConn(lUser, addr)
}

func closeConn(lUser, addr string) {
	if _, ok := userMap[lUser]; ok {
		if len(userMap[lUser]) <= 1 {
			//close the connection
			userMap[lUser][0].Conn.Close()
			delete(userMap, lUser)
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
}

//tell the guy that the opposite guy has leaved and he can connect to another guy
func reportTUserOffLine(lUser, addr string) {
	var chatter string
	var conn net.Conn
	if _, ok := userMap[lUser]; ok {
		for i := 0; i < len(userMap[lUser]); i++ {
			if userMap[lUser][i].Addr == addr {
				chatter = userMap[lUser][i].chatter
				break
			}
		}
	}

	if chatter != "" {
		if _, ok := userMap[chatter]; ok {
			for i := 0; i < len(userMap[chatter]); i++ {
				if userMap[chatter][i].chatter == lUser {
					userMap[chatter][i].chatter = ""
					conn = userMap[chatter][i].Conn
					break
				}
			}
		}
	}

	if conn != nil {
		msg := fmt.Sprintf("%s已下线\n%s", lUser, remindMsg2)

		util.Write(conn, msg)
	}
}

//when a login user connect to a waiting guy, then force buiding the connection
func forceConnection(lUser, chatter, addr string) (chatterAddr string, err error) {
	//var isExistWaiting bool
	if us, ok := userMap[chatter]; ok {
		for i := 0; i < len(us); i++ {
			if us[i].chatter == "" {
				chatterAddr = us[i].Addr
				us[i].chatter = lUser
				us[i].ChatterAddr = addr
				util.Write(us[i].Conn, fmt.Sprintf(remindMsg3, lUser))
				//isExistWaiting = true
				return
			}
		}
	}
	return "", errors.New(remindMsg4)
}
