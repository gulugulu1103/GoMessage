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
	flag       int // 当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil

	}

	client.conn = conn

	// 返回对象
	return client
}

// menu displays the menu options for the client and prompts the user to select an option.
// It reads the input from the user and validates if the input is within the range of available options.
// If the input is valid, it returns true. Otherwise, it displays an error message and returns false.
func (client *Client) menu() bool {
	var choice int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&choice)
	if err != nil {
		return false
	}

	if choice >= 0 && choice <= 3 {
		client.flag = choice
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<")
		return false
	}
}

func (client *Client) Rename() {
	fmt.Println(">>>>>请输入用户名：")
	_, err := fmt.Scanln(&client.Name)
	if err != nil {
		return
	}

	sendMsg := "rename " + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (client *Client) SecretChat() {
	fmt.Println(">>>>>请输入聊天对象用户名：")
	var chatName string
	_, err := fmt.Scanln(&chatName)
	if err != nil {
		return
	}

	fmt.Println(">>>>>请输入聊天内容：")
	var chatMsg string
	_, err = fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}

	sendMsg := "msg " + chatName + " " + chatMsg + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (client *Client) PublicChat() {
	fmt.Println(">>>>>请输入聊天内容：")
	var chatMsg string
	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}

	sendMsg := chatMsg + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
	}

	switch client.flag {
	case 1: // 公聊模式
		client.PublicChat()
		break
	case 2: // 私聊模式
		client.SecretChat()
		break
	case 3: // 更新用户名
		client.Rename()
		break
	}

}

// DealResponse 用于处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		return
	}

	for {
		// 接收server发送的消息
		buffer := make([]byte, 4096) // 4k大小的缓冲区
		// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
		_, err := client.conn.Read(buffer)
		if err != nil {
			return
		}
		fmt.Println(string(buffer))
	}
}

var serverIp string
var serverPort int

// init sets the initial values of the server IP and port.
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

// main is the entry point for the application.
// It parses command line arguments and initializes a client.
// If the client fails to initialize, it prints an error message and exits.
// Otherwise, it prints a success message and starts the client's business logic.
// The function then enters a select statement, which blocks the main goroutine and keeps the program running.
func main() {
	flag.Parse() // 解析命令行参数

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>链接服务器失败...")
		return
	}

	fmt.Println(">>>>>链接服务器成功...")

	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse() // 处理server回应的消息

	// 启动客户端的业务
	client.Run()
}
