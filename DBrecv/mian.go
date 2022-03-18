package main

import (
	"DBRECV/mylog"
	"DBRECV/proto"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var (
	Host   string
	Port   int
	QpPath string
)

// 定义结果结构体
type Result struct {
	Retcode int64
	Size    int64
	Msg     string
}

// 处理命令行参数
func HandleArgs(host *string, port *int, qpPath *string) error {
	flag.StringVar(host, "host", "0.0.0.0", "bind host")
	flag.IntVar(port, "port", -1, "bind port")
	flag.StringVar(qpPath, "qp-path", "", "qp file path")
	flag.Parse()
	if *port == -1 {
		mylog.Log.Error("port must be gave in command")
		flag.Usage()
		return fmt.Errorf("port must be gave in command")
	}
	if *qpPath == "" {
		mylog.Log.Error("qppath must be gave in command")
		flag.Usage()
		return fmt.Errorf("qppath must be gave in command")

	}
	return nil
}

// RecvStream 接收客户端传递的数据写入qp文件
func RecvStream(conn net.Conn, qpPath string) (int64, error) {
	mylog.Log.Info(fmt.Sprintf("%v has connected in", conn.RemoteAddr()))
	var (
		readSize  int64
		writeSize int64
	)

	// 创建qp文件
	f, err := os.OpenFile(qpPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		return 0, fmt.Errorf(fmt.Sprintf("create file %s failed, err: %s", qpPath, err))
	}
	defer func() {
		// 关闭资源
		f.Close()
		err := recover() // 重现错误
		if err != nil {
			panic(err)
		}
	}()
	mylog.Log.Info(fmt.Sprintf("create qp file:%v succeed", qpPath))

	var head [4]byte
	for {
		// 读取头部
		if _, err := conn.Read(head[:4]); err != nil {
			return 0, fmt.Errorf(fmt.Sprintf("read data head error, err: %s", err))
		}
		lengthBuff := bytes.NewBuffer(head[:4])
		var length int32
		if err := binary.Read(lengthBuff, binary.LittleEndian, &length); err != nil {
			return 0, fmt.Errorf(fmt.Sprintf("trans head to length error, err: %s", err))
		}
		dataLength := int64(length)
		fmt.Println("data length -> ", dataLength)
		if dataLength < 0 {
			return 0, fmt.Errorf("read lenth error")
		}

		// 读取实际数据
		var data = make([]byte, dataLength)
		n, err := conn.Read(data[:])
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf(fmt.Sprintf("read from client failed, err: %s", err))
		}

		if n == 2 {
			if data := string(data); data == "OK" {
				mylog.Log.Info(fmt.Sprintf("recv stream data from %s over", conn.RemoteAddr()))
				break
			}
		}

		readSize += int64(n)

		writeLength, err := f.Write(data)

		if err != nil {
			return 0, fmt.Errorf(fmt.Sprintf("write to qppath:%s failed, err:%s", qpPath, err))
		}
		writeSize += int64(writeLength)
	}
	if readSize == writeSize {
		mylog.Log.Info(fmt.Sprintf("write data to %s succeed, data size is %d bytes", qpPath, writeSize))
		return readSize, nil
	}
	return 0, fmt.Errorf("read data size is not equal to write data")
}

func main() {
	// 初始化LOG
	mylog.InitLogger(Port)
	// 获取命令行参数
	if err := HandleArgs(&Host, &Port, &QpPath); err != nil {
		panic(err)
	}
	// 建立连接
	listen, err := net.Listen("tcp", fmt.Sprintf("%v:%d", Host, Port))
	if err != nil {
		mylog.Log.Error("listen failed, err:", err)
		panic(fmt.Sprintf("listen failed, err: %s", err))
	}

	mylog.Log.Info(fmt.Sprintf("stream recv server run, bind: %v:%d", Host, Port))

	conn, err := listen.Accept() // 建立连接
	if err != nil {
		mylog.Log.Error("accept failed, err:", err)
	}
	var recvRet Result
	defer func() {
		conn.Close()
		listen.Close()
		e := recover()
		if e != nil {
			mylog.Log.Error(fmt.Sprintf("recv stream from %s failed, err:%s", conn.RemoteAddr(), e))
		}
	}()
	readSize, err := RecvStream(conn, QpPath)

	if err != nil {
		mylog.Log.Error(err)
		recvRet.Retcode = 400
	}
	// 赋值
	recvRet.Retcode = 200
	recvRet.Size = readSize

	// 回传结果
	ret, _ := json.Marshal(recvRet)
	ret, err = proto.Encode(ret)
	if err != nil {
		mylog.Log.Error("marshal recv result failed, err:%s", err)
	}
	if _, err = conn.Write(ret); err != nil {
		mylog.Log.Error("send recv result failed, err:", err)
		panic(fmt.Sprintf("send recv result failed, err:%s", err))
	}

	// 接收客户端回传的对比结果

	ResultReader := bufio.NewReader(conn)
	result, err := proto.Decode(ResultReader)
	if err != nil {
		mylog.Log.Error(fmt.Sprintf("get send stream result from %s, err: ", conn.RemoteAddr()), err)
		panic(fmt.Sprintf("get send stream result from %s, err: %s", conn.RemoteAddr(), err))
	}

	var sendResult Result

	err = json.Unmarshal(result, &sendResult)

	if err != nil {
		mylog.Log.Error("decode failed, err:", err)
		panic(fmt.Sprintf("decode failed, err:%s", err))
	}
	if sendResult.Retcode == 200 {
		mylog.Log.Info(fmt.Sprintf("recv stream from %s succeed", conn.RemoteAddr()))
	}

}
