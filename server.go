package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个Server对象
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 当前链接的业务
func (s *Server) Handler(conn net.Conn) {
	fmt.Println("Links build successfully")
	// 用户上线
	user := NewUser(conn, s)
	// 封装功能
	user.Online()
	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 当前handler阻塞
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				// return退出当前线程，而不退出Handler
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 获取用户的信息
			msg := string(buf[:n-1])
			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()

	// 超时处理，定时器
	for {
		select {
		case <-isLive:
			// 只要能从isLive中获取数据，则说明用户在活跃，重启定时器
		case <-time.After(time.Second * 10):
			// 一但进入该case就说明超时了，After本质是一个channel，当超时时会向管道写入标记
			// 超时时将当前User强制关闭
			user.SendMsg("You have been quit.")
			// 销毁资源，只关闭user.C
			close(user.C)
			return // 或使用runtime.Goexit()
		}
	}
}

// 广播信息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		// 将msg发送给全部的在线User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.listener err:", err)
		return
	}

	// close socket listener
	defer listener.Close()
	// 启动监听Message的goroutine
	go s.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		s.Handler(conn)
	}
}
