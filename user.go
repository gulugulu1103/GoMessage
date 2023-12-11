package main

import (
	"net"
	"strings"
)

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
	} else if len(msg) >= 7 && msg[:7] == "rename " { // 消息格式：rename 新名字
		newName := msg[7:]
		if newName == "" {
			this.SendMessage("名字错误，请重试，格式：rename 新名字\n")
			return
		}
		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMessage("当前用户名被使用\n")
		} else {
			// 更新用户名
			this.server.mapLock.Lock()
			// 将用户从OnlineMap中删除
			delete(this.server.OnlineMap, this.Name)
			// 将用户添加到OnlineMap中
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMessage("您已经更新用户名：" + this.Name + "\n")
		}
	} else if len(msg) >= 4 && msg[:4] == "msg " { // 消息格式：msg 名字 消息内容
		// 检查消息格式是否正确
		if len(strings.Split(msg, " ")) < 3 {
			this.SendMessage("消息格式不正确，请使用\"msg 用户名 消息内容\"格式\n")
			return
		}
		// 1. 获取对方的用户名
		toName := strings.Split(msg, " ")[1]
		if toName == "" {
			this.SendMessage("消息格式不正确，请使用\"msg 用户名 消息内容\"格式\n")
			return
		}
		// 2. 根据用户名得到对方User对象
		toUser, ok := this.server.OnlineMap[toName] // toUser是一个指针，指向对方的User对象
		if !ok {
			this.SendMessage("该用户名不存在\n")
			return
		}
		// 3. 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, " ")[2]
		if content == "" {
			this.SendMessage("无消息内容，请重发，格式：msg 用户名 消息内容\n")
			return
		}
		// 4. 发送
		this.SendMessage("你对" + toUser.Name + "说：" + content + "\n")
		toUser.SendMessage(this.Name + "对你说：" + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}
