package proxy

import (
	"io"
	"log"
	"net"
	"time"
)

var 	Cfg *PorxyConfig

func connCopy2(dst io.Writer, src io.Reader, stopChan chan bool ) {
	io.Copy(dst, src)
	stopChan <- true
}

func connCopy(dst net.Conn, src net.Conn, stopChan chan bool ) {
	var err error
	var read int
	bytes := make([]byte, 10240)
	for   {
		dst.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
		read, err = dst.Read(bytes)
		if err != nil {
			break
		}
		_, err = src.Write(bytes[:read])
		if err != nil {
			break
		}
	}
	stopChan <- true
}

func handleConn(client net.Conn, backend string,stopChan chan bool) {
	defer func() {
		BackendSvrs[backend].Connections--
	}()
	BackendSvrs[backend].Connections++
	server, err := net.DialTimeout("tcp", backend,  time.Duration(10)*time.Second)
	if err != nil {
		log.Println(err)
		client.Close()
		return
	}
	connStopChan := make(chan bool,2)
	go connCopy(server, client, connStopChan)
	go connCopy(client, server, connStopChan)
	select {
	case <- connStopChan:
		server.Close()
		client.Close()
		<- connStopChan
		log.Println("close connection",backend,client.RemoteAddr())
	case <- stopChan:
		server.Close()
		client.Close()
		<- connStopChan
		<- connStopChan
		log.Println("kill connection",backend,client.RemoteAddr())
	}

}

func KillConn(num int, stopChan chan bool) {
	for i := 1; i <= num; i++ {
		stopChan<-true
	}
}

func StartPorxy(config *PorxyConfig) {
	l, err := net.Listen("tcp", ":"+config.Listen)
	if err != nil {
		log.Panic(err)
	}
	Cfg = config
	backendSvrs := GetBackendSvrs(config)
	ws := WeightNext(backendSvrs)

	go StartCheckHealth()
	go ManageStart(config.Mport)

	for {
		client, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			break
		}
		svr,stopChan,err := GetNextBackendSvr(ws)
		if err != nil {
			log.Println("backend error:", err)
			client.Close()
			continue
		}
		go handleConn(client, svr, stopChan)
	}
}