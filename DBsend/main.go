package main

import (
	"DBsend/proto"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

const CHUNKSIZE = 4092

// ip地址格式钩子
func VarIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	return ip, 1

}

var (
	Host string
	Port int
)

// 处理命令行参数
func HandleArgs(host *string, port *int) error {
	flag.StringVar(host, "host", "0.0.0.0", "bind host")
	flag.IntVar(port, "port", -1, "bind port")
	flag.Parse()
	if *port == -1 {
		flag.Usage()
		return fmt.Errorf("port must be gave in command")
	}
	return nil
}

// 定义结果结构体
type Result struct {
	Retcode int64
	Size    int64
	Msg     string
}

// 客户端
func main() {
	// 接收命令行参数
	err := HandleArgs(&Host, &Port)
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%d", Host, Port))
	if err != nil {
		panic(fmt.Sprintf("connect to %s:%d failed, err: %s", Host, Port, err))
	}
	defer func() {
		conn.Close() // 关闭资源
		e := recover()
		if e != nil {
			panic(e)
		}
	}()

	var (
		recvRet  Result
		sendSize int64
	)

	reader := bufio.NewReader(os.Stdin)

	var buf [CHUNKSIZE]byte
	// 传输数据
	for {
		n, err := reader.Read(buf[:CHUNKSIZE]) // 读取stdin
		if err != nil && err != io.EOF {
			panic(fmt.Sprintf("read data from stdin failed, err: %s \n", err))
		}
		if err == io.EOF || n == 0 {
			OK, _ := proto.Encode([]byte("OK"))
			conn.Write(OK)
			break
		}
		data, err := proto.Encode(buf[:n])
		if err != nil {
			panic(fmt.Sprintf("encode data failed, err: %s \n", err))
		}
		conn.Write(data) // 发送数据
		sendSize += int64(n)
	}
	// 接收回传
	retReader := bufio.NewReader(conn)
	recvResult, err := proto.Decode(retReader)
	// 判断服务端接收的数据是否与发送数据大小一致
	if err != nil {
		fmt.Println("error")
		return
	}
	if err := json.Unmarshal(recvResult, &recvRet); err != nil{
		panic(fmt.Sprintf("unmarshal data error, err: %s", err))
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
		if err != nil {
			panic(fmt.Sprintf("send data error, err: %s", err))
		}
	} else {
		panic("send stream failed")
	}

}
