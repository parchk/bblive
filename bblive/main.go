// mss project main.go
package main

import (
	"bbllive/conf"
	"bbllive/log"
	"bbllive/rtmp"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

var pprof bool
var listen string

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	//flag.StringVar(&listen, "l", ":9993", "-l=:1935")
	//flag.BoolVar(&pprof, "pprof", false, "-pprof=true")

	//listen = conf.AppConf.RtmpAddress
	//pprof = conf.AppConf.PProf
}

func main() {

	log.Debugf("kkk %d,%s", int32(10), "ssssssss")
	flag.Parse()
	err := rtmp.ListenAndServe(conf.AppConf.RtmpAddress)
	if err != nil {
		panic(err)
	}

	log.Debug("rtmp ListenAndServe :", conf.AppConf.RtmpAddress)

	if conf.AppConf.PProf {
		go func() {
			log.Debug(http.ListenAndServe(":6060", nil))
		}()
	}

	log.Debug("ListenAndServerHttpFlv :", conf.AppConf.HttpFlvAddress)

	herr := rtmp.ListenAndServerHttpFlv(conf.AppConf.HttpFlvAddress)

	if herr != nil {
		log.Error("ListenAndServerHttpFlv error :", herr)
		return
	}

	select {}
}
