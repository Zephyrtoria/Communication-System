package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
	// 绑定Server，以便能够访问到server
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听
	go user.ListenMessage()

	return user
}

// 用户上线功能
func (user *User) Online() {
	// 用户上线，将用户加入到OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 广播当前用户上线信息
	user.server.BroadCast(user, "已上线")
}

// 用户下线功能
func (user *User) Offline() {
	// 用户下线，将用户从OnlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	// 广播当前用户上线信息
	user.server.BroadCast(user, "已下线")
}

// 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	// 可以有其他的处理操作，这里恰好只有调用广播接口的功能
	user.server.BroadCast(user, msg)
}

// 监听当前User channel的方法，一旦有消息就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
