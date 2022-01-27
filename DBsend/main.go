package main

import (
	"DBsend/proto"
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

const CHUNKSIZE = 2048

// ip地址格式钩子
func VarIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	return ip, 1

}

// 处理命令行参数
func HandleArgs() (string, int) {
	var host string
	var port int
	flag.StringVar(&host, "host", "0.0.0.0", "bind host")
	flag.IntVar(&port, "port", 20001, "bind port")
	flag.Parse()

	return host, port
}

// 客户端
func main() {
	// 接收命令行参数
	host, port := HandleArgs()
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		fmt.Println("Connect err:", err)
		return
	}
	defer conn.Close() // 关闭连接
	reader := bufio.NewReader(os.Stdin)
	sendSize := 0
	var buf [CHUNKSIZE]byte
	// 传输数据
	for {
		n, err := reader.Read(buf[:]) // 读取stdin

		if err != nil {
			fmt.Println("read err", err)
		}
		// 制作数据包
		message, err := proto.Encode(buf[:])
		if err != nil {
			fmt.Println("encode error", err)
		}
		_, err = conn.Write(message) // 发送数据
		if err != nil {
			return
		}
		sendSize += n
	}
	// 接收回传
	data, err := proto.Decode(reader)
	// 判断服务端接收的数据是否与发送数据大小一致
	if err != nil {
		fmt.Println("error")
		return
	}
	recvSize, err := strconv.Atoi(data)
	if err != nil {
		fmt.Println("error")
		return
	}
	if recvSize == sendSize {
		// 传输结束信号
		flag := true
		// 将flag转换成[]byte
		flagMsg, err := proto.Encode(flag)
	}

}
