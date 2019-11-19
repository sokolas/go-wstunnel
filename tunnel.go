package main
import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	// "net/url"
	// "github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	localPort := flag.Int("lport", 1488, "local port to listen")
	remoteAddr := flag.String("remote", "localhost:1488", "remote address:port to connect to")

	wsAddr := flag.String("ws", "ws://localhost:1080/ws", "websocket URL")

	flag.Parse()
	fmt.Println(*localPort)
	fmt.Println(*remoteAddr)
	fmt.Println(*wsAddr)


	signal.Notify(interrupt, syscall.SIGINT)
	go func() {
		sig := <- interrupt
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("Waiting...")
	<- done
	fmt.Println("Done")
}