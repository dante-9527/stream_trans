package main

import (
	"fmt"
	"net"
	"bufio"
	"os"
)

// 定制log
func setLogger(){
	fmt.Println("定制log")
}


// 处理函数
func process(conn net.Conn) {
	defer conn.Close() // 关闭连接
	read_size := 0
	write_size := 0
	for {
		fmt.Printf("%s has connected in", conn)
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(os.Stdout)
		var buf [1024]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		read_size += n
		if n == 0 && err != nil{
			fmt.Println("read over")
		}
		nn, err := writer.Write(buf[:n])
		if err != nil {
			fmt.Println("write to stdout failed, err", err)
		}
		write_size += nn
		if read_size == write_size{
			fmt.Println("读写结束")
			break
		}
	}
}

func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}
	for {
		conn, err := listen.Accept() // 建立连接
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}
		go process(conn) // 启动一个goroutine处理连接
	}
}
