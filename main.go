package main

func main() {
	server := Server{
		Ip:   "127.0.0.1",
		Port: 8888,
	}
	server.Start()
}
