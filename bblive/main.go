// mss project main.go
package main

import (
	"bbllive/conf"
	"bbllive/log"
	"bbllive/rtmp"
	"flag"
	"net"
	"net/http"
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
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

func lookPath() (argv0 string, err error) {
	argv0, err = exec.LookPath(os.Args[0])
	if nil != err {
		return
	}
	if _, err = os.Stat(argv0); nil != err {
		return
	}
	return
}

func main() {

	log.Debugf("kkk %d,%s", int32(10), "ssssssss")
	flag.Parse()

	go func() {
		defer func() {
			if x := recover(); x != nil {
				log.Info("run time panic: %v", x)
			}
			log.Error("run time panic")
		}()

		herr := rtmp.ListenAndServerHttpFlv(conf.AppConf.HttpFlvAddress)

		if herr != nil {
			log.Error("ListenAndServerHttpFlv error :", herr)
			return
		}
	}()

	err := rtmp.ListenAndServe(conf.AppConf.RtmpAddress)
	if err != nil {
		panic(err)
	}

	log.Debug("rtmp ListenAndServe :", conf.AppConf.RtmpAddress)

	if conf.AppConf.PProf {
		go func() {
			addr_s, err := net.ResolveTCPAddr("tcp", ":6060")

			if err != nil {
				log.Error("rtmp Listen resolve  addr error:", err)
				return
			}

			l, err := rtmp.Gracennet_net.ListenTCP("tcp", addr_s)

			if err != nil {
				log.Error("rtmp Listen pprof 6060 error:", err)
			}

			err = http.Serve(l, nil)

			if err != nil {
				log.Error("rtmp pprof server end error :", err)
			}
		}()
	}

	log.Debug("ListenAndServerHttpFlv :", conf.AppConf.HttpFlvAddress)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)

	for sig := range signals {

		if sig == syscall.SIGTERM {
			// Stop accepting new connections
			log.Debug("Server shutdown successful")

			os.Exit(0)

		} else if sig == syscall.SIGHUP {
			/*
				_, file, _, err := rtmp.GetListenFD()

				if err != nil {
					log.Error("Fail to get socket file descriptor:", err)
				}

				_, web_file, _, err := rtmp.GetHttpFlvListFD()

				if err != nil {
					log.Error("httpflv Fail to get socket file descriptor:", err)
				}

				// Set a flag for the new process start process

				os.Setenv("GRACEFUL_RESTART", "true")

				_, err, _ = rtmp.SetEnvs()
				_, err, _ = rtmp.SetPocWebEvn()

				if err != nil {
					log.Error("fork set event error :", err)
					os.Exit(1)
				}

				//log.Info("httpfd :", httpfd, httpaddr.Network(), httpaddr.String())
				//log.Info("fd :", fd, addr.Network(), addr.String())
				/*
					var max_fd uintptr

					if web_fd > fd {
						max_fd = web_fd
					} else {
						max_fd = fd
					}
			*/
			/*
				allFiles := make([]*os.File, max_fd+1)
				allFiles[syscall.Stdin] = os.Stdin
				allFiles[syscall.Stdout] = os.Stdout
				allFiles[syscall.Stderr] = os.Stderr

					allFiles[fd] = os.NewFile(
						fd,
						fmt.Sprintf("%s:%s->", addr.Network(), addr.String()),
					)
					allFiles[httpfd] = os.NewFile(
						httpfd,
						fmt.Sprintf("%s:%s->", httpaddr.Network(), httpaddr.String()),
					)
			*/
			/*
				//files := make([]uintptr, max_fd+1)

				//files[os.Stdin.Fd()] = os.Stdin.Fd()
				//files[os.Stdout.Fd()] = os.Stdout.Fd()
				//files[os.Stderr.Fd()] = os.Stderr.Fd()
				//files[fd] = fd
				//files[web_fd] = web_fd

					execSpec := &syscall.ProcAttr{
						Env: os.Environ(),
						//Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), fd, web_fd},
						Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), fd, web_fd},
					}

					// Fork exec the new version of your server
					fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
					if err != nil {
						log.Error("Fail to fork", err)
					}
			*/
			/*
				argv0, err := lookPath()
				if nil != err {
					log.Error("lookPath error:", err)
					os.Exit(1)
				}
				wd, err := os.Getwd()
				if nil != err {
					log.Error("os GetWd error:", err)
					os.Exit(1)
				}

				p, err := os.StartProcess(argv0, os.Args, &os.ProcAttr{
					Dir:   wd,
					Env:   os.Environ(),
					Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, file, web_file},
				})

				log.Info("SIGHUP received: fork-exec to", p.Pid)
				// Wait for all conections to be finished
			*/
			rtmp.Gracennet_net.StartProcess()
			rtmp.StopListen()
			rtmp.StopWebListen()
			rtmp.WaitStop()

			log.Info(os.Getpid(), "Server gracefully shutdown")

			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}
