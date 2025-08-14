package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/zh-five/xdaemon"
)

// 初始化UDP连接
func initUDPConnection(ip string, port string) (*net.UDPConn, *net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on UDP: %v", err)
	}

	return conn, addr, nil
}

// 处理UDP通信
func handleUDPCommunication(conn *net.UDPConn, addr *net.UDPAddr, verbose bool) error {
	buffer := make([]byte, 1024)
	n, clientAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		// 忽略所有连接关闭相关的错误
		if strings.Contains(err.Error(), "use of closed network connection") || err == net.ErrClosed {
			return nil
		}
		
		if verbose {
			log.Printf("Error reading from UDP: %v", err)
		}
		// 检查是否是连接断开错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("UDP connection timeout, reconnecting...")
			// 重新创建连接
			newConn, err := net.ListenUDP("udp", addr)
			if err != nil {
				return fmt.Errorf("failed to reconnect UDP: %v", err)
			}
			*conn = *newConn
		}
		return err
	}

	if verbose {
		log.Printf("Received message from %s: %s", clientAddr, string(buffer[:n]))
	}

	// 限制返回数据长度为8字节
	returnData := buffer[:n]
	if len(returnData) > 8 {
		returnData = returnData[:8]
	}

	// 返回截取后的数据
	_, err = conn.WriteToUDP(returnData, clientAddr)
	if err != nil {
		if verbose {
			log.Printf("Error writing to UDP: %v", err)
		}
		// 检查是否是连接断开错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("UDP write timeout, reconnecting...")
			// 重新创建连接
			newConn, err := net.ListenUDP("udp", addr)
			if err != nil {
				return fmt.Errorf("failed to reconnect UDP: %v", err)
			}
			*conn = *newConn
		}
		return err
	}

	return nil
}

func main() {
	// 定义命令行参数
	ip := flag.String("i", "0.0.0.0", "IP address to listen on")
	port := flag.String("p", "23832", "Port to listen on")
	daemon := flag.Bool("d", false, "Run as a daemon")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()
	//启动守护进程
	if *daemon {
		//创建一个Daemon对象
		logFile := "daemon.log"
		daemon := xdaemon.NewDaemon(logFile)
		//调整一些运行参数(可选)
		daemon.MaxCount = 2 //最大重启次数
		//执行守护进程模式
		daemon.Run()
	}

	// 初始化UDP连接
	conn, addr, err := initUDPConnection(*ip, *port)
	if err != nil {
		log.Fatalf("Failed to initialize UDP connection: %v", err)
	}
	defer conn.Close()

	if *verbose {
		log.Printf("UDP server is running on %s:%s\n", *ip, *port)
	}

	// 处理信号，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		conn.Close()
		os.Exit(0)
	}()
	// 主循环
	for {
		if err := handleUDPCommunication(conn, addr, *verbose); err != nil {
			log.Printf("Error handling UDP communication: %v", err)
		}
	}
}
