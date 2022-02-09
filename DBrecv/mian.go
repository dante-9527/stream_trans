package main

import (
	"DBRECV/mylog"
	"DBRECV/proto"
	"bufio"
	"encoding/json"
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
	flag.IntVar(&port, "port", 20001, "bind port")
	flag.StringVar(&qpPath, "qp-path", "", "qp file path")

	flag.Parse()

	return host, port, qpPath
}

// 处理函数
func process(conn net.Conn, qpPath string) (int, error){
	mylog.Log.Info("%s has connected in", conn)
	fmt.Println(conn)
	readSize := 0
	writeSize := 0
	// 创建qp文件
	f, err := os.OpenFile(qpPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	defer func() {
		// 关闭资源
		f.Close()
		err := recover() // 重现错误
		if err != nil{
			panic(err)
		}
	}()
	if err != nil {
		return 0, fmt.Errorf(fmt.Sprintf("create file %s failed, err: %s", qpPath, err))
	}
	mylog.Log.Info(fmt.Sprintf("create qpfile:%s succeed",qpPath))
	
	for {
		reader := bufio.NewReader(conn)
		// 读取数据
		data, err := proto.Decode(reader)
		if err != nil {
			return 0, fmt.Errorf(fmt.Sprintf("read from client failed, err: %s", err))
		}
		if data == nil && err == nil {
			mylog.Log.Info(fmt.Sprintf("recv stream data from %s over", conn))
			break
		}
		readLength := len(data)
		
		readSize += readLength

		writeLength, err := f.Write(data)
		if err != nil {
			return 0, fmt.Errorf(fmt.Sprintf("write to qppath:%s failed, err:%s", qpPath, err))
		}
		writeSize += writeLength
	}
	if readSize == writeSize {
		mylog.Log.Info(fmt.Sprintf("write data to %s succeed, data size is %d bytes", qpPath, writeSize))
		f.Close()
		return readSize, nil
	}
	return 0, fmt.Errorf("read data size is not equal to write data")
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
		panic(fmt.Sprintf("listen failed, err: %s", err))
	}

	mylog.Log.Info(fmt.Sprintf("stream recv server run, bind: %v:%d", host, port))

	conn, err := listen.Accept() // 建立连接
	if err != nil {
		mylog.Log.Error("accept failed, err:", err)
	}
	// 声明结果结构体
	type Result struct {
		Retcode int
		Size    int
		Msg     string
	}
	var recvRet Result
	defer func() {
		// 回传结果
		fmt.Println(recvRet)
		ret, _ := json.Marshal(recvRet)
		ret, err = proto.Encode(ret)
		fmt.Println(ret)
		if err != nil{
			mylog.Log.Error("marshal recv result failed, err:%s", err)
		}
		_, err = conn.Write(ret) // 发送数据
		if err != nil {
			mylog.Log.Error("send recv result failed, err:", err)
			panic(fmt.Sprintf("send recv result failed, err:%s", err))
		}

		// 接收客户端回传的对比结果
		var result []byte
		
		var sendResult Result
		for {
			ResultReader := bufio.NewReader(conn)
			ret, err := proto.Decode(ResultReader)
			if err != nil {
				mylog.Log.Error(fmt.Sprintf("get send stream result from %s, err: ", conn), err)
				panic(fmt.Sprintf("get send stream result from %s, err: %s", conn, err))
			}
			if ret == nil {
				break
			}
			result = append(result, ret...)
		}
		fmt.Println(result)
		err = json.Unmarshal(result, &sendResult)
		if err != nil {
			mylog.Log.Error("decode failed, err:", err)
			panic(fmt.Sprintf("decode failed, err:%s", err))
		}
		if sendResult.Retcode == 200{
			mylog.Log.Info(fmt.Sprintf("recv stream from %s succed", conn))
			conn.Close()
			listen.Close()
		}
		e := recover()
		if e != nil{
			mylog.Log.Error(fmt.Sprintf("recv stream from %s failed, err:%s", conn, err))
		}
	}()
	readSize, err := process(conn, qpPath) // 处理连接
	
	if err != nil{
		mylog.Log.Error(err)
		recvRet.Retcode = 400
		panic(err)
	}
	// 赋值
	recvRet.Retcode = 200
	recvRet.Size = readSize
	
}
