package main
import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"net"
	// "net/url"
	// "github.com/gorilla/websocket"
)

func handleConnection(from net.Conn) {
	defer from.Close()
	fmt.Println("New connection accepted", from.RemoteAddr())
	//to, err := net.Dial("tcp", *remoteAddr) - TODO dial websocket
	// if err != nil {
		// handle error
	// }
	// run copy
}

func main() {
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	localPort := flag.Int("lport", 1488, "local port to listen")
	remoteAddr := flag.String("remote", "localhost:1488", "remote address:port to connect to")
	wsAddr := flag.String("ws", "ws://localhost:1080/ws", "websocket URL")

	flag.Parse()
	fmt.Println("Local port", *localPort)
	fmt.Println("Remote address", *remoteAddr)
	fmt.Println("WebSocket URL", *wsAddr)


	signal.Notify(interrupt, syscall.SIGINT)
	go func() {
		sig := <- interrupt
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()


	ln, err := net.Listen("tcp", "127.0.0.1:" + fmt.Sprintf("%d", *localPort))
	defer ln.Close()
	if err != nil {
		fmt.Println("Listening error", err)
		return
	}
	
	// accepted := make(chan *Conn, 1)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
				fmt.Println("Listener:", err)
				return
			}

			go handleConnection(conn)
		}
	}()



	fmt.Println("Ready")
	<- done
	fmt.Println("Done")
}