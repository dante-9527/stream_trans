package main

import(
	"fmt"
	"net"
	"bufio"
	"os"
	"flag"
)

// ip地址格式钩子
func VarIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	return ip,1
	
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
	host, port := HandleArgs()
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%d",host,port))
	if err != nil {
		fmt.Println("Connect err:", err)
		return
	}
	defer conn.Close() // 关闭连接
	inputReader := bufio.NewReader(os.Stdin)
	sendSize := 0
	var buf [1024]byte
	for {
		_, err := inputReader.Read(buf[:]) // 读取stdin
		if err != nil{
			fmt.Println("read err", err)
		}
		n, err := conn.Write(buf[:]) // 发送数据
		if err != nil {
			return
		}
		sendSize += n
		
	}
	
}