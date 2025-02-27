package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// 创建一个Server对象
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

// 当前链接的业务
func (s *Server) Handler(conn net.Conn) {
	fmt.Println("链接建立成功")
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
