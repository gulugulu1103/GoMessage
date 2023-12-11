package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

// Server 服务器结构体
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

// Handler 处理当前连接的业务
func (this *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	//fmt.Println("链接建立成功")

	// 用户上线，将用户加入到OnlineMap中
	user := NewUser(conn, this)

	user.Online() // 用户上线
	isLive := make(chan bool)
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
				user.Offline() // 用户下线
				return
			}

			// 提取用户的消息（去除'\n'）
			msg := string(buffer[:n-1])
			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			isLive <- true // 证明当前用户是活跃的

		}
	}()

	for {
		select {
		case <-isLive: // 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(60 * time.Second): // time.After是一个channel，60秒后，自动往channel中写内容，此时select就可以执行了
			// 超时退出
			user.SendMessage("你已超时，已被踢出")
			// 销毁用的资源
			close(user.C)
			// 关闭连接
			err := conn.Close()
			if err != nil { // 关闭失败
				fmt.Println("conn.Close err:", err)
				return
			}

			runtime.Goexit() // 结束当前的goroutine

		}
	}
}

// Start 启动服务器的接口
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
