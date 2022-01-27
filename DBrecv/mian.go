package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

// 定制log
func setLogger() {
	fmt.Println("定制log")
}

// 处理命令行参数
func HandleArgs() (string, int, string) {
	var host string
	var port int
	var qpPath string
	flag.StringVar(&host, "host", "0.0.0.0", "bind host")
	flag.IntVar(&port, "port", 20001, "bind port")
	flag.StringVar(&qpPath, "qp-path", "", "qp file path")

	flag.Parse()

	return host, port, qpPath
}

// 处理函数
func process(conn net.Conn, qpPath string) {
	defer conn.Close() // 关闭连接
	fmt.Printf("%s has connected in", conn)
	readSize := 0
	writeSize := 0
	// 创建qp文件
	f, err := os.Create(qpPath)
	defer f.Close()
	if err != nil {
		fmt.Println("err=", err)
		return
	}
	for {
		reader := bufio.NewReader(conn)
		var buf [1024]byte
		readLength, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		readSize += readLength
		if readLength == 0 && err != nil {
			fmt.Println("read over")
			break
		}
		writeLength, err := f.Write(buf[:])
		if err != nil {
			fmt.Println("write to qppath failed, err:", err)
			break
		}
		writeSize += writeLength
	}
	if readSize == writeSize{
		fmt.Println("handler success")
	}
}

func main() {
	// 获取命令行参数
	host, port, qpPath := HandleArgs()
	// 建立连接
	listen, err := net.Listen("tcp", fmt.Sprintf("%v:%d", host, port))
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
		go process(conn, qpPath) // 启动一个goroutine处理连接
	}
}
