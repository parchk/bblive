// mss project main.go
package main

import (
	"bbllive/conf"
	"bbllive/log"
	"bbllive/rtmp"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	go func() {
		herr := rtmp.ListenAndServerHttpFlv(conf.AppConf.HttpFlvAddress)

		if herr != nil {
			log.Error("ListenAndServerHttpFlv error :", herr)
			return
		}
	}()
	log.Debug("ListenAndServerHttpFlv :", conf.AppConf.HttpFlvAddress)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)

	for sig := range signals {

		if sig == syscall.SIGTERM {
			// Stop accepting new connections
			log.Debug("Server shutdown successful")

			os.Exit(0)

		} else if sig == syscall.SIGHUP {

			_, err := rtmp.GetListenFD()

			if err != nil {
				log.Error("Fail to get socket file descriptor:", err)
			}

			// Set a flag for the new process start process

			//syscall.CloseOnExec(int(listenerFD))

			allFiles := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}

			/*
				for _, i := range rtmp.GetConFile() {
					log.Debug("111111111111")
					//syscall.CloseOnExec(int(i.Fd()))
					allFiles = append(allFiles, i.Fd())
				}
			*/

			os.Setenv("GRACEFUL_RESTART", "true")

			rtmp.SetEnvs()
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: allFiles,
			}

			// Fork exec the new version of your server
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Error("Fail to fork", err)
			}

			log.Info("SIGHUP received: fork-exec to", fork)
			// Wait for all conections to be finished

			rtmp.WaitStop()

			log.Info(os.Getpid(), "Server gracefully shutdown")

			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}
