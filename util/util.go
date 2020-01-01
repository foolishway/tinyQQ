package util

import (
	"bufio"
	//"bytes"
	//"fmt"
	//"io"
	"net"
	"strings"
)

func Write(conn net.Conn, content string) (int, error) {
	writer := bufio.NewWriter(conn)
	number, err := writer.WriteString(strings.TrimSpace(content))
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}
func Read(conn net.Conn) (content string, err error) {
	//reader := bufio.NewReader(conn)
	//var buffer bytes.Buffer
	//for {
	//	ba, isPrefix, err := reader.ReadLine()
	//	if err != nil {
			//if err == io.EOF {
	//		//	fmt.Println("util read error (EOF): %s", err)
	//		//	break
	//		//}
	//		//fmt.Printf("util read error: %s", err)
	//		return "", err
	//	}
	//	buffer.Write(ba)
	//	if !isPrefix {
	//		break
	//	}
	//}
	//return buffer.String(), err
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	return string(buf[:n]), err
}
