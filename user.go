package main

import (
	"net"
	"strings"
)

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
	user.server.BroadCast(user, "is online")
}

// 用户下线功能
func (user *User) Offline() {
	// 广播当前用户下线信息
	user.server.BroadCast(user, "is offline")
	// 将用户从OnlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

}

// 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	// 查询指令
	if msg == "who" {
		// 查询当前在线用户
		user.server.mapLock.Lock()
		for _, each := range user.server.OnlineMap {
			onlineMsg := "[" + each.Addr + "]" + each.Name + ":" + "is currently online\n"
			// 给指定用户发送消息
			user.conn.Write([]byte(onlineMsg))
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[0:7] == "rename|" {
		// 消息格式 rename|newName
		newName := strings.Split(msg, "|")[1]
		// 判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("This username is used.")
		} else {
			user.server.mapLock.Lock()
			// 删除当前名字
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("User name update:" + newName)
		}
	} else {
		user.server.BroadCast(user, msg)
	}
}

// 监听当前User channel的方法，一旦有消息就直接发送给对端客户端
func (user *User) ListenMessage() {
	for msg := range user.C {
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			panic(err)
		}
	}
	err := user.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (user *User) SendMsg(msg string) {
	user.C <- msg
}
