package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bind, listen, accept
	ln, err := net.Listen("tcp", ":8181")
	if err != nil {
		log.Fatalf("announces on the local network address: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalf("wait for the incoming call: %v", err)
			}

			t := newTony(ctx, conn)
			go t.read()
			go t.write()
		}
	}()

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		println("Received SIGTERM, exiting gracefully...")
	}
	println("Bye!!!")
}

// tony .
type tony struct {
	ctx  context.Context
	conn net.Conn
	ch   chan []byte
}

func newTony(ctx context.Context, conn net.Conn) *tony {
	return &tony{
		ctx:  ctx,
		conn: conn,
		ch:   make(chan []byte),
	}
}

func (t *tony) read() {
	defer t.conn.Close()

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			content := make([]byte, 1024)
			_, err := t.conn.Read(content)
			if err != nil {
				println("read content: " + err.Error())
				return
			}
			println("read content: " + string(content))
			t.ch <- content
		}
	}
}

func (t *tony) write() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case content := <-t.ch:
			_, err := t.conn.Write(content)
			if err != nil {
				println("write content: " + err.Error())
				return
			}
			println("write content: " + string(content))
		default:
		}
	}
}
