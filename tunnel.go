package main
import (
	"flag"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net"
	"net/http"
	"net/url"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

func handleConnection(tcp net.Conn, u string) {
	defer tcp.Close()
	fmt.Println("New connection accepted", tcp.RemoteAddr())

	header := make(http.Header)
	header["x-wstclient"] = []string{uuid.New().String()}
	header["Sec-Websocket-Protocol"] = []string{"tunnel-protocol"}

	ws, _, error := websocket.DefaultDialer.Dial(u, header)
	if error != nil {
		log.Println("dial:", error)
		return
	}
	defer ws.Close()

	done := make(chan interface{})
	defer close(done)
	// make 2 goroutines, one for tcp reading, one for ws reading
	// both write to done on exit/error
	go func() {
		for {
			_, bytes, err := ws.ReadMessage()
			if err != nil {
				log.Println("ws read error:", err)
				done <- err
				return
			}
			n, err := tcp.Write(bytes)
			if err != nil {
				log.Println("tcp write error:", err)
				done <- err
				return
			}
			if n != len(bytes) {
				err = errors.New("Failed to write all data")
				log.Println("tcp rwite error:", err)
				done <- err
				return
			}
		}
	}()

	go func() {
		bytes := make([]byte, 2048)
		for {
			n, err := tcp.Read(bytes)
			if err != nil {
				log.Println("tcp read error:", err)
				done <- err
				return
			}
			err = ws.WriteMessage(websocket.BinaryMessage, bytes[:n])
			if err != nil {
				log.Println("ws write error:", err)
				done <- err
				return
			}
		}
	}()
	err1, err2 := <- done, <- done
	log.Println("Done with", tcp.RemoteAddr())
	log.Println(err1, err2)
	// run deferred closes
}

func main() {
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	localPort := flag.Int("lport", 1488, "local port to listen")
	remoteAddr := flag.String("remote", "localhost:1488", "remote address:port to connect to")
	wsAddr := flag.String("ws", "ws://localhost:8000/ws", "websocket URL")

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

			go handleConnection(conn, *wsAddr + "/?dst=" + *remoteAddr)
		}
	}()



	log.Println("Ready")
	<- done
	log.Println("Done")
}