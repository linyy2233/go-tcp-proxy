package main

import (
	"flag"
	"github.com/linyy2233/go-tcp-proxy/proxy"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
)

func main() {

	var config proxy.PorxyConfig
	var configFile = flag.String("c", "config.yaml", "配置文件，默认config.yaml")
	var cpuprofile = flag.String("cpuprofile", "proxy.cpuprofile", "write cpu profile to this file")
	var memprofile = flag.String("memprofile", "proxy.memprofile", "write memory profile to this file")

	flag.Parse()
	log.Println("Start Proxy...")
	log.Println(*configFile)
	conf := config.GetConf(*configFile)
	go proxy.StartPorxy(conf)

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR1)

		s := <-c
		log.Println("Got signal:", s)

		memf, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(memf)
		memf.Close()

		cpuf, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(cpuf); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		time.Sleep(10 * time.Second)
		pprof.StopCPUProfile()
		cpuf.Close()
}
