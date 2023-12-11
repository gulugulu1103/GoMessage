package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server // 当前用户所属的Server，用于后续退出的消息广播
}

// NewUser 用户的构造函数
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage() // 启动

	return user
}

// ListenMessage 监听自己的UserChannel，一旦有消息，就发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		_, err := this.conn.Write([]byte(msg + "\n")) // 发送消息给客户端
		if err != nil {
			return
		}
	}

}

// Online 用户上线，将用户加入到OnlineMap中
func (this *User) Online() {
	// 用户上线，将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")

}

// Offline 用户下线，将用户从OnlineMap中删除
func (this *User) Offline() {
	// 用户下线，将用户从OnlineMap中删除
	this.server.mapLock.Lock()
	// 将用户从OnlineMap中删除
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已下线")
}

// SendMessage 给当前用户对应的客户端发送消息
func (this *User) SendMessage(msg string) {
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}

// DoMessage 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "list" {
		// 查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMessage(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.BroadCast(this, msg)
	}
}
