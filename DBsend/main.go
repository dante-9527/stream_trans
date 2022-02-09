package main

import (
	"DBsend/proto"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

const CHUNKSIZE = 1024

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
	// 定义结果结构体
	type Result struct {
		Retcode int
		Size    int
		Msg     string
	}
	var recvRet Result
	if err != nil {
		panic(fmt.Sprintf("connect to %s:%d failed, err: %s", host, port, err))
	}
	defer func() {
		conn.Close() // 关闭资源
		e := recover()
		if e != nil {
			panic(e)
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	sendSize := 0
	var buf [CHUNKSIZE]byte
	// 传输数据
	for {
		n, err := reader.Read(buf[:]) // 读取stdin
		fmt.Println("n------", n, err)
		if n != 0 && err != nil {
			panic(fmt.Sprintf("read data from stdin failed, err: %s", err))
		}
		// 制作数据包
		data, err := proto.Encode(buf[:n])
		if err != nil {
			panic(fmt.Sprintf("encode data error, err: %s", err))
		}
		nn, err := conn.Write(data) // 发送数据
		if err != nil {
			panic(fmt.Sprintf("send data error, err: %s", err))
		}
		fmt.Println(nn)
		if n == 0 {
			break
		}
		sendSize += n
	}
	// 接收回传
	fmt.Println("111111111")
	retReader := bufio.NewReader(conn)
	recvResult, err := proto.Decode(retReader)
	fmt.Println(recvResult)
	// 判断服务端接收的数据是否与发送数据大小一致
	if err != nil {
		fmt.Println("error")
		return
	}
	e := json.Unmarshal(recvResult, &recvRet)
	
	if e != nil {
		panic(fmt.Sprintf("unmarshal data error, err: %s", e))
	}
	if recvRet.Retcode == 200 {
		var judgeRet Result
		if recvRet.Size == sendSize {
			judgeRet.Retcode = 200
		} else {
			judgeRet.Retcode = 400
		}
		retData, _ := json.Marshal(judgeRet)
		retData, err = proto.Encode(retData)
		if err != nil {
			panic(fmt.Sprintf("encode data error, err: %s", err))
		}
		_, err = conn.Write(retData) // 发送结果
		conn.Close()
		if err != nil {
			panic(fmt.Sprintf("send data error, err: %s", err))
		}
	} else {
		panic("send stream failed")
	}

}
