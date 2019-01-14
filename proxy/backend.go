package proxy

import (
	"errors"
	"github.com/smallnest/weighted"
	"log"
	"net"
	"time"
)

type BackendSvr struct {
	SvrStr    string
	IsUp      bool
	Weight int
	CheckTimeout int
	CheckFail int
	CheckInter time.Duration
	ActiveTries int
	InactiveTries int
	IsStop bool
	Connections int
	StopChan chan bool
	MaxConn int
}

var (
	BackendSvrs map[string]*BackendSvr
)

func GetBackendSvrs(config *PorxyConfig) map[string]*BackendSvr {
	BackendSvrs = make(map[string]*BackendSvr)

	for _, svr := range config.BackendCfg.Backends {
		stopChan := make(chan bool)
		BackendSvrs[svr.Addr] = &BackendSvr{
			SvrStr:    svr.Addr,
			IsUp:      true,
			Weight: svr.Weight,
			CheckFail: config.BackendCfg.CheckFail,
			CheckInter: config.BackendCfg.CheckInter,
			CheckTimeout: config.BackendCfg.CheckTimeout,
			ActiveTries: 0,
			InactiveTries: 0,
			IsStop: false,
			StopChan: stopChan,
			MaxConn: svr.MaxConn,
			Connections: 0,
		}
	}
	return BackendSvrs
}

func WeightNext(backendSvrs map[string]*BackendSvr) (*weighted.SW) {
	w := &weighted.SW{}
	for _,v := range backendSvrs {
		w.Add(v.SvrStr,v.Weight)
	}
	return w
}

func GetNextBackendSvr(w *weighted.SW) (string,chan bool, error) {
	failNum := 0
	for _,b := range BackendSvrs {
		if !b.IsUp || b.IsStop {
			failNum++
		}
	}
	if failNum == len(BackendSvrs) {
		return "null",nil,errors.New("503 service unavailable")
	}
	var svr string
	for  {
		wss := w.Next()
		if v, ok := wss.(string); ok {  // checked type assertion
			svr = v
		}
		if BackendSvrs[svr].IsUp && !BackendSvrs[svr].IsStop && BackendSvrs[svr].Connections < BackendSvrs[svr].MaxConn {
			break
		}
	}
	return svr ,BackendSvrs[svr].StopChan,nil
}

func StartCheckHealth()  {
	for _,b := range BackendSvrs {
		go func(b *BackendSvr) {
			for   {
				b.InactiveTries = 0
				b.ActiveTries = 0
				for i:=0;i<=b.CheckFail;i++ {
					time.Sleep(b.CheckInter * time.Second)
					if b.IsStop {
						continue
					}
					if b.ActiveTries >= b.CheckFail && !b.IsUp {
						b.IsUp = true
						b.InactiveTries = 0
						b.ActiveTries = 0
						log.Println(b.SvrStr, "is up")
					} else if b.IsUp && b.InactiveTries >= b.CheckFail {
						b.IsUp = false
						b.ActiveTries = 0
						log.Println(b.SvrStr, "is down")
					} else {
						conn, err := net.DialTimeout("tcp", b.SvrStr, time.Duration(b.CheckTimeout)*time.Second)
						if err != nil {
						//	log.Println(err)
							b.InactiveTries++
						} else {
							conn.Close()
							b.ActiveTries++
						}
					}
				}
			}
		}(b)
	}
}

