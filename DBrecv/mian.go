package main

import (
	"DBRECV/mylog"
	"DBRECV/proto"
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

// 处理命令行参数
func HandleArgs() (string, int, string) {
	var host string
	var port int
	var qpPath string
	flag.StringVar(&host, "host", "0.0.0.0", "bind host")
	flag.IntVar(&port, "port", 0, "bind port")
	flag.StringVar(&qpPath, "qp-path", "", "qp file path")

	flag.Parse()

	return host, port, qpPath
}

// 处理函数
func process(conn net.Conn, qpPath string) {
	defer conn.Close() // 关闭连接
	mylog.Log.Info("%s has connected in", conn)
	readSize := 0
	writeSize := 0
	// 创建qp文件
	f, err := os.OpenFile(qpPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("err=", err)
		// f.Close()
		return
	}
	for {
		reader := bufio.NewReader(conn)
		// 读取数据
		data, err := proto.Decode(reader)
		if err != nil {
			mylog.Log.Error("Read from client failed, err:", err)
			break
		}
		readLength := len(data)
		if readLength == 0 && err != nil {
			mylog.Log.Info(fmt.Sprintf("Recv stream data from %s over", conn))
			break
		}
		readSize += readLength

		writeLength, err := f.Write(data)
		if err != nil {
			mylog.Log.Error(fmt.Sprintf("Write to qppath:%s failed, err:%s", qpPath, err))
			break
		}
		writeSize += writeLength
	}
	if readSize == writeSize {
		mylog.Log.Info(fmt.Sprintf("Write data to %s succeed, data size is %d bytes", qpPath, writeSize))
		f.Close()
	}
}

func main() {
	// 初始化LOG
	mylog.InitLogger()
	// 获取命令行参数
	host, port, qpPath := HandleArgs()
	// 建立连接
	listen, err := net.Listen("tcp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		mylog.Log.Error("listen failed, err:", err)
		return
	}
	mylog.Log.Info(fmt.Sprintf("Stream recv server run, bind: %v:%d", host, port))
	for {
		conn, err := listen.Accept() // 建立连接
		if err != nil {
			mylog.Log.Error("accept failed, err:", err)
			continue
		}
		go process(conn, qpPath) // 启动一个goroutine处理连接
	}
}
