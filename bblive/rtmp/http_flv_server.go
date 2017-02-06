package rtmp

import (
	"bbllive/conf"
	"bbllive/log"
	"fmt"
	"net"
	_ "net"
	"net/http"
	"os"
	"reflect"
	"syscall"
	"time"

	_ "github.com/rcrowley/goagain"
)

type HttpFlvPlayHandle struct {
	l net.Listener
}

func (*HttpFlvPlayHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println("request ParseForm error")
		http.NotFound(w, r)
		return
	}

	var obj *StreamObject

	id := r.Form.Get("s")

	objname := fmt.Sprintf("live/%s", id)

	obj, found := FindObject(objname)

	if !found {

		t := time.After(time.Second * 3)

		<-t

		obj, found = FindObject(objname)

		if !found {

			log.Error("object not find key :", objname)
			http.NotFound(w, r)
			return
		}
	}

	srv.Wp.Add(1)

	stream := NewHttpFlvStream(objname)
	stream.SetObj(obj)

	obj.HttpAttach(stream)

	log.Debug("HttpFlvStream BeginHanle")

	if conf.AppConf.GOPCache {
		stream.WriteLoop(w, r)
	} else {
		log.Debug("HttpFlvStream begin write loopf")
		stream.WriteLoopF(w, r)
	}

	log.Debug("HttpFlvStream mid")

	stream.Close()

	log.Debug("HttpFlvStream EndHanle")

	log.Info("(((((((stream close")

	srv.Wp.Done()

}

func SetPocWebEvn(l net.Listener) (err error) {

	v := reflect.ValueOf(l).Elem().FieldByName("fd").Elem()

	fd := uintptr(v.FieldByName("sysfd").Int())

	_, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFD, 0)

	if 0 != e1 {
		err = e1
		return
	}

	if err = os.Setenv("GOAGAIN_FD_WEB", fmt.Sprint(fd)); nil != err {
		return
	}
	addr := l.Addr()

	if err = os.Setenv(
		"GOAGAIN_NAME_WEB",
		fmt.Sprintf("%s:%s->", addr.Network(), addr.String()),
	); nil != err {
		return
	}

	return
}

func ListenFromFD() (l net.Listener, err error) {

	var fd uintptr

	if _, err = fmt.Sscan(os.Getenv("GOAGAIN_FD_WEB"), &fd); nil != err {
		return nil, err
	}

	l, err = net.FileListener(os.NewFile(fd, os.Getenv("GOAGAIN_NAME_WEB")))

	if nil != err {
		return nil, err
	}

	switch l.(type) {

	case *net.TCPListener, *net.UnixListener:
	default:
		err = fmt.Errorf(
			"file descriptor is %T not *net.TCPListener or *net.UnixListener",
			l,
		)
		return nil, err
	}

	if err = syscall.Close(int(fd)); nil != err {
		return nil, err
	}

	return l, nil
}

func ListenFD(addr string) (net.Listener, error) {

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	return l, nil
}

func ListenAndServerHttpFlv(addr string) error {

	mux := http.NewServeMux()
	handle := &HttpFlvPlayHandle{}
	mux.Handle("/play", handle)
	server := &http.Server{Addr: addr, Handler: mux}

	var err error
	var l net.Listener

	if os.Getenv("GRACEFUL_RESTART") == "true" {
		l, err = ListenFromFD()
	} else {
		l, err = ListenFD(addr)
	}

	SetPocWebEvn(l)

	if err != nil {
		return err
	}

	return server.Serve(l)
}
