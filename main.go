package main

import (
  "fmt"
  "os"
  "os/signal"
	"log"
  "flag"
  "net/http"
  "code.google.com/p/go.net/websocket"
)

type connection struct {
  ws *websocket.Conn
  send chan string
}

func (c *connection) reader() {
  for {
    var message string
    err := websocket.Message.Receive(c.ws, &message)
    if err != nil {
      break
    }
    h.broadcast <- message
  }
  c.ws.Close()
}

func (c *connection) writer() {
  for message := range c.send {
    log.Print(message)
    err := websocket.Message.Send(c.ws, message)
    if err != nil {
      break
    }
  }
  c.ws.Close()
}

func wsHandler(ws *websocket.Conn) {
  c := &connection{send: make(chan string, 256), ws: ws}
  h.register <- c
  defer func() { h.unregister <- c }()
  go c.writer()
  c.reader()
}

type hub struct {
  connections map[*connection]bool
  broadcast chan string
  register chan *connection
  unregister chan *connection
}

var h = hub{
  broadcast: make(chan string),
  register: make(chan *connection),
  unregister: make(chan *connection),
  connections: make(map[*connection]bool),
}

func (h *hub) run() {
  for {
    select {
      case c:= <-h.register:
        h.connections[c] = true
        log.Print("Connection registered", c)
      case c:= <-h.unregister:
        delete(h.connections, c)
        close(c.send)
      case m:= <-h.broadcast:
        for c:= range h.connections {
          select {
          case c.send <- m:
          default:
            delete(h.connections, c)
            close(c.send)
            go c.ws.Close()
        }
      }
    }
  }
}

var addr = flag.String("addr", ":8080", "http service address")

func listen() {
  http.Handle("/", websocket.Handler(wsHandler))

  log.Print("Server starting")
  if err := http.ListenAndServe(*addr, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}

func main() {
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)

  go func() {
		for _ = range c {
			// sig is a ^C, handle it
			log.Fatal("Finished - bye bye")
		}
	}()

  flag.Parse()
  go h.run()
  go listen()
  for {
    var input string
    _, err := fmt.Scanf("%s", &input)
    if err != nil {
      log.Print(err)
    }
    h.broadcast <- input
  }
}

