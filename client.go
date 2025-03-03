package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

var serverIp string
var serverPort int

func init() {
	// 设置命令行参数
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认为127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器Port地址（默认为8888）")
}

func NewClient(serverIp string, serverPort int) *Client {

	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       1,
	}

	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err: ", err)
		return nil
	}

	client.conn = conn
	return client
}

func (client *Client) menu() bool {
	fmt.Println("1. Public")
	fmt.Println("2. Private")
	fmt.Println("3. Rename")
	fmt.Println("0. Quit")

	var flag int
	fmt.Scanln(&flag)
	if 0 <= flag && flag <= 3 {
		client.flag = flag
		return true
	}
	fmt.Println("Invalid input.")
	return false
}

func (client *Client) Run() {
	for client.menu() && client.flag != 0 {
		switch client.flag {
		case 1:
			// 公聊模式
			fmt.Println("Public talk mode")
			client.PublicTalk()
			break
		case 2:
			// 私聊模式
			fmt.Println("Private talk mode")
			client.PrivateTalk()
			break
		case 3:
			// 更新用户名
			fmt.Println("Rename mode")
			client.UpdateName()
			break
		}
	}
}

// menu功能实现
func (client *Client) PublicTalk() {
	// 提示输入消息
	var chatMsg string

	fmt.Println("Please input message, input 'exit' to quit")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Message send fail")
				break
			}
		}
		fmt.Println("Please input message, input 'exit' to quit")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) PrivateTalk() {
	// 查询当前用户
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("Online user look up failed")
		return
	}

	// 提示输入消息
	var chatMsg string
	var targetUser string
	fmt.Println("Please input target username")
	fmt.Scanln(&targetUser)
	fmt.Println("Please input message, input 'exit' to quit")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := "to|" + targetUser + "|" + chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Message send fail")
				break
			}
		}
		fmt.Println("Please input message, input 'exit' to quit")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("Input your new name")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("Update name error")
		return false
	}
	return true
}

// 处理server回应的消息，单独使用一个go来运行
func (client *Client) DealResponse() {
	// 一但client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	// 单独开启一个goroutine去处理服务器发送过来的消息
	go client.DealResponse()
	fmt.Println("连接服务器成功！")

	// 启动客户端业务
	client.Run()
}
