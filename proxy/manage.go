package proxy

import (
	"github.com/toolkits/web/param"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)


func ManageStart(port string) {

	http.HandleFunc("/hello", HelloServer)
	http.HandleFunc("/down", DownBackend)
	http.HandleFunc("/up", UpBackend)
	http.HandleFunc("/status", StatusBackend)
	http.HandleFunc("/close", CloseBackend)
	http.HandleFunc("/dispatch", DispatchBackend)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func StatusBackend(w http.ResponseWriter, r *http.Request) {
	var resultAll string
	for k,v  := range BackendSvrs {
		conn := strconv.Itoa(v.Connections)
		if v.IsStop {
			resultAll = resultAll + k + " stopped"
		} else {
			resultAll = resultAll + k + " running"
		}
		if v.IsUp {
			resultAll = resultAll  + " " + conn + " up \n"
		} else {
			resultAll = resultAll  + " " + conn + " down \n"
		}
	}
	io.WriteString(w, resultAll)
}

func DownBackend(w http.ResponseWriter, r *http.Request)  {
	var resultAll string
	svrString := param.MustString(r, "svr")
	token := param.String(r, "configtoken","")
	svrList := strings.Split(svrString, ",")

	if token == Cfg.ConfigToken {
		for _,svr := range svrList {
			if _, ok := BackendSvrs[svr]; !ok {
				resultAll := svr + " not found\n"
				http.Error(w,  resultAll, 400)
				return
			}
		}

		for _,svr := range svrList {
			BackendSvrs[svr].IsStop = true
			KillConn(BackendSvrs[svr].Connections,BackendSvrs[svr].StopChan)
		}

	} else {
		for _,proxy := range Cfg.ProxyList {
			data := make(url.Values)
			data["configtoken"] = []string{Cfg.ConfigToken}
			data["svr"] = []string{svrString}
			requestUrl := "http://"+proxy+"/down"
			res, err := http.PostForm(requestUrl, data)
			if err != nil {
				log.Println(err)
				resultAll = resultAll + requestUrl + " " +svrString+ " failed \n"
				continue
			}
			result, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(result))
			resultAll = resultAll + requestUrl + " " +svrString+ " "+string(result) +"\n"
		}
	}

	http.Error(w,  resultAll, http.StatusOK)
}

func UpBackend(w http.ResponseWriter, r *http.Request)  {
	var resultAll string
	svrString := param.MustString(r, "svr")
	token := param.String(r, "configtoken","")
	svrList := strings.Split(svrString, ",")

	if token == Cfg.ConfigToken {
		for _,svr := range svrList {
			if _, ok := BackendSvrs[svr]; !ok {
				resultAll := svr + " not found\n"
				http.Error(w,  resultAll, 400)
				return
			}
		}
		for _,svr := range svrList {
			BackendSvrs[svr].IsStop = false
		}

	} else {
		for _,proxy := range Cfg.ProxyList {
			data := make(url.Values)
			data["configtoken"] = []string{Cfg.ConfigToken}
			data["svr"] = []string{svrString}
			requestUrl := "http://"+proxy+"/up"
			res, err := http.PostForm(requestUrl, data)
			if err != nil {
				log.Println(err)
				resultAll = resultAll + requestUrl + " " +svrString+ " failed \n"
				continue
			}
			result, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(result))
			resultAll = resultAll + requestUrl + " " +svrString+" "+ string(result) +"\n"
		}
	}

	http.Error(w,  resultAll, http.StatusOK)
}

func CloseBackend(w http.ResponseWriter, r *http.Request)  {
	var resultAll string
	svrString := param.MustString(r, "svr")
	ratio := param.MustInt(r, "ratio")
	token := param.String(r, "configtoken","")
	svrList := strings.Split(svrString, ",")

	if token == Cfg.ConfigToken {
		for _,svr := range svrList {
			if _, ok := BackendSvrs[svr]; !ok {
				resultAll := svr + " not found\n"
				http.Error(w,  resultAll, 400)
				return
			}
		}
		for _,svr := range svrList {
			killNum := BackendSvrs[svr].Connections / ratio
			go KillConn(killNum,BackendSvrs[svr].StopChan)
		}

	} else {
		for _,proxy := range Cfg.ProxyList {
			data := make(url.Values)
			data["configtoken"] = []string{Cfg.ConfigToken}
			data["svr"] = []string{svrString}
			data["ratio"] = []string{strconv.Itoa(ratio)}
			requestUrl := "http://"+proxy+"/close"
			res, err := http.PostForm(requestUrl, data)
			if err != nil {
				log.Println(err)
				resultAll = resultAll + requestUrl + " " +svrString+ " failed \n"
				continue
			}
			result, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(result))
			resultAll = resultAll + requestUrl + " " +svrString+" " + string(result) + "\n"
		}
	}

	http.Error(w,  resultAll, http.StatusOK)
}

func DispatchBackend(w http.ResponseWriter, r *http.Request)  {
	var resultAll string
	fromString := param.MustString(r, "from")
	toString := param.MustString(r, "to")
	ratio := param.MustInt(r, "ratio")
	token := param.String(r, "configtoken","")
	fromList := strings.Split(fromString, ",")
	toList := strings.Split(toString, ",")

	if token == Cfg.ConfigToken {
		for _,svr := range fromList {
			if _, ok := BackendSvrs[svr]; !ok {
				resultAll := svr + " not found\n"
				http.Error(w,  resultAll, 400)
				return
			}
		}
		for _,svr := range toList {
			if _, ok := BackendSvrs[svr]; !ok {
				resultAll := svr + " not found\n"
				http.Error(w,  resultAll, 400)
				return
			}
		}
		for _,b := range BackendSvrs {
			b.IsUp = false
			b.ActiveTries = 0
		}
		go func() {
			for i := 0; i < 20; i++ {
				time.Sleep(3 * time.Second)
				for _,b := range BackendSvrs {
					b.IsUp = false
					b.ActiveTries = 0
				}
				for _,svr := range toList {
					BackendSvrs[svr].IsUp = true
				}
			}
		}()
		for _,svr := range toList {
			BackendSvrs[svr].IsUp = true
		}
		for _,svr := range fromList {
			killNum := BackendSvrs[svr].Connections / ratio
			go KillConn(killNum,BackendSvrs[svr].StopChan)
		}

	} else {
		for _,proxy := range Cfg.ProxyList {
			data := make(url.Values)
			data["configtoken"] = []string{Cfg.ConfigToken}
			data["to"] = []string{toString}
			data["from"] = []string{fromString}
			data["ratio"] = []string{strconv.Itoa(ratio)}
			requestUrl := "http://"+proxy+"/dispatch"
			res, err := http.PostForm(requestUrl, data)
			if err != nil {
				log.Println(err)
				resultAll = resultAll + requestUrl + " from " +fromString+  " to " +toString+ " failed \n"
				continue
			}
			result, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(result))
			resultAll = resultAll + requestUrl + " from " +fromString+  " to " +toString+ " " +string(result) + "\n"
		}
	}

	http.Error(w,  resultAll, http.StatusOK)
}
