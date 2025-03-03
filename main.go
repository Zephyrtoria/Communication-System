package main

// server.go和main.go同属于main包，不需要import

func main() {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
