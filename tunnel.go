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

func sendWsClose(ws websocket.Conn) error {
	err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("ws sendClose error: ", err)
	}
	return err
}

func handleConnection(tcp net.Conn, u string) {
	addr := tcp.RemoteAddr()
	log.Println("New connection accepted", addr)

	header := make(http.Header)
	header["x-wstclient"] = []string{uuid.New().String()}
	header["Sec-Websocket-Protocol"] = []string{"tunnel-protocol"}

	ws, _, error := websocket.DefaultDialer.Dial(u, header)
	if error != nil {
		log.Println("dial:", error)
		return
	}

	doneTcp := make(chan interface{})
	doneWs := make(chan interface{})
	// make 2 goroutines, one for tcp reading, one for ws reading
	// both write to done on exit/error
	go func() {
		defer log.Println(addr, "Finished ws read")
		for {
			_, bytes, err := ws.ReadMessage()
			if err != nil {
				log.Println(addr, "ws read:", err)
				doneWs <- err
				return
			}
			n, err := tcp.Write(bytes)
			if err != nil {
				log.Println(addr, "tcp write:", err)
				doneWs <- err
				return
			}
			if n != len(bytes) {
				err = errors.New("Failed to write all data")
				log.Println(addr, "tcp write:", err)
				doneWs <- err
				return
			}
		}
	}()

	go func() {
		defer log.Println(addr, "Finished tcp read")
		bytes := make([]byte, 2048)
		for {
			n, err := tcp.Read(bytes)
			if err != nil {
				log.Println(addr, "tcp read:", err)
				doneTcp <- err
				return
			}
			err = ws.WriteMessage(websocket.BinaryMessage, bytes[:n])
			if err != nil {
				log.Println(addr, "ws write:", err)
				doneTcp <- err
				return
			}
		}
	}()
	
	for i := 0; i < 2; i++ {
		select {
		case err1 := <- doneTcp:
			log.Println(addr, "tcp reader: ", err1)
			_ = sendWsClose(*ws)
			ws.Close()
		case err2 := <- doneWs:
			log.Println(addr, "ws reader: ", err2)
			tcp.Close()
		}
	}
	log.Println("Done with", tcp.RemoteAddr())
}

func main() {
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool)

	localPort := flag.Int("lport", 1489, "local port to listen")
	remoteAddr := flag.String("remote", "localhost:1489", "remote address:port to connect to")
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
