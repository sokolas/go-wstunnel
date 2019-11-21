package main
import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net"
	"net/url"
	// "github.com/gorilla/websocket"
)

func handleConnection(from net.Conn, u string) {

	defer from.Close()
	fmt.Println("New connection accepted", from.RemoteAddr())
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan error)
	defer close(done)
	// make 2 goroutines, one for tcp reading, one for ws reading
	// both write to done on exit/error

	err1, err2 := <- done, <- done

}

func main() {
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	localPort := flag.Int("lport", 1488, "local port to listen")
	remoteAddr := flag.String("remote", "localhost:1488", "remote address:port to connect to")
	wsAddr := flag.String("ws", "ws://localhost:1080/ws", "websocket URL")

	flag.Parse()
	log.Println("Local port", *localPort)
	log.Println("Remote address", *remoteAddr)

	u, err := url.Parse(*wsAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("WebSocket URL", u)

	
	signal.Notify(interrupt, syscall.SIGINT)
	go func() {
		sig := <- interrupt
		log.Println(sig)
		done <- true
	}()


	ln, err := net.Listen("tcp", "127.0.0.1:" + fmt.Sprintf("%d", *localPort))
	defer ln.Close()
	if err != nil {
		log.Println(err)
		return
	}
	
	// accepted := make(chan *Conn, 1)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
				log.Println(err)
				return
			}

			go handleConnection(conn, *wsAddr + "/?dst=" + remoteAddr)
		}
	}()



	log.Println("Ready")
	<- done
	log.Println("Done")
}