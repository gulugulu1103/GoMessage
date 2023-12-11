package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的Channel
	Message chan string
}

// NewServer 创建一个Server的接口
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// ListenMessage 监听Message广播消息Channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast ListenMessage 监听Message广播消息Channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg

}

func (this *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	//fmt.Println("链接建立成功")

	// 用户上线，将用户加入到OnlineMap中
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	this.BroadCast(user, "已上线")

	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := conn.Read(buffer)
			// 读取用户发送的消息，Read方法是读取客户端发送的数据，如果客户端没有write，那么就会阻塞在这里
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			// 提取用户的消息（去除'\n'）
			msg := string(buffer[:n-1])

			this.BroadCast(user, msg)
		}
	}()

	select {}
}

func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// defer 尝试关闭listener
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close err:", err)
			return
		}
	}(listener)

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		go this.Handler(conn)

	}
	// accept

	// do handler
	// close
}
