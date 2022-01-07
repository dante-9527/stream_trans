package main

import(
	"fmt"
	"net"
	"bufio"
	"os"
)


// 客户端
func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println("Connect err:", err)
		return
	}
	defer conn.Close() // 关闭连接
	inputReader := bufio.NewReader(os.Stdin)
	var buf [1024]byte
	for {
		n, err := inputReader.Read(buf[:]) // 读取stdin
		if err != nil{
			fmt.Println("read err", err)
		}
		_, err = conn.Write(buf[:n]) // 发送数据
		if err != nil {
			return
		}
		fmt.Println(string(buf[:n]))
	}
}