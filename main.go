package main

import (
  "os"
  "os/signal"
  "net"
  "net/http"
	"log"
  "time"
  "encoding/json"
)

const (
  Address string = ":80"
)

func main() {
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  listener, err := net.Listen("tcp", Address)
}
