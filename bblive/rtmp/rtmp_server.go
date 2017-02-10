package rtmp

import (
	"bbllive/conf"
	"bbllive/log"
	_ "bbllive/util"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/facebookgo/grace/gracenet"
	_ "github.com/sdming/gosnow"
	cmap "github.com/streamrail/concurrent-map"
)

var Gracennet_net = &gracenet.Net{}

var (
	objects = cmap.New()
	//log      *util.FileLogger
	shandler ServerHandler = new(DefaultServerHandler)
	logfile  string
	level    int
	//snow     *gosnow.SnowFlake
	srvid int

	srv *Server
)

type NotifyObj struct {
	Lock sync.RWMutex
	Lc   map[string]chan string
}

func (n *NotifyObj) WaitObj(objname string) bool {
	var cn chan string

	NObje.Lock.RLock()

	ocn, ok := NObje.Lc[objname]

	if ok {

		cn = ocn

	} else {

		cn = make(chan string)

		NObje.Lc[objname] = cn
	}

	NObje.Lock.RUnlock()

ForEnd:
	for {
		select {
		case <-cn:
			NObje.Lock.Lock()
			delete(NObje.Lc, objname)
			NObje.Lock.Unlock()
			break ForEnd
		case <-time.After(time.Second * 2):
			log.Debug("waite object time out :", objname)
			return false
		}
	}

	return true
}

func (n *NotifyObj) NotifyObje(objname string) {

	NObje.Lock.RLock()
	defer NObje.Lock.RUnlock()

	cn, ok := NObje.Lc[objname]

	if ok {
		select {
		case cn <- objname:
		case <-time.After(time.Millisecond * 500):
			log.Info("OnPublishing notify httpflv obj time out :", objname)
		}
	}
}

var NObje NotifyObj

func init() {
	/*
		flag.StringVar(&logfile, "log", "stdout", "-log rtmp.log")
		flag.IntVar(&level, "level", 1, "-level 1")
		flag.IntVar(&srvid, "srvid", 1, "-srvid 1")
	*/
	logfile = conf.AppConf.LogPath
	level = conf.AppConf.LogLvl
	srvid = conf.AppConf.Srvid

	stdlog.SetFlags(stdlog.Lmicroseconds | stdlog.Lshortfile)
	stdlog.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))

	NObje.Lc = make(map[string]chan string)
}

func Timer() {

	for {
		select {
		case <-time.After(10 * time.Second):
			objects.IterCb(func(key string, v interface{}) {
				log.Info("object key :", key, " value :", v)
			})
		}
	}
}

func WaitStop() {

	log.Info("rtmp server watie $$$$$$")
	srv.Wp.Wait()

}

func GetListenFD() (uintptr, *os.File, net.Addr, error) {

	file, err := srv.l.File()

	if err != nil {
		return 0, nil, nil, err
	}

	return file.Fd(), file, srv.l.Addr(), nil
}

func SetEnvs() (uintptr, error, []string) {
	return srv.setEnvs()
}

func GetConFile() []*os.File {
	return srv.files
}

func ListenAndServe(addr string) error {
	/*
		logger, err := util.NewFileLogger("", logfile, level)
		if err != nil {
			return err
		}
	*/
	//log = logger
	/*
		snow, err = gosnow.NewSnowFlake(uint32(srvid))
		if err != nil {
			return err
		}
	*/
	srv = &Server{
		Addr:         addr,
		ReadTimeout:  time.Duration(time.Second * 30),
		WriteTimeout: time.Duration(time.Second * 30),
		Lock:         new(sync.Mutex),
		Wp:           new(sync.WaitGroup)}

	//go Timer()

	return srv.ListenAndServe()
}

func StopListen() error {
	err := srv.l.Close()
	srv.CloseFlag = true
	return err
}

type Server struct {
	Addr         string        //监听地址
	ReadTimeout  time.Duration //读超时
	WriteTimeout time.Duration //写超时
	Lock         *sync.Mutex
	Wp           *sync.WaitGroup
	l            *net.TCPListener
	files        []*os.File
	CloseFlag    bool
}

/*
func nsid() int {
	id, _ := conf.Snow.Next()
	return int(id)
}
*/
var gstreamid = uint32(64)

func gen_next_stream_id(chunkid uint32) uint32 {
	gstreamid += 1
	return gstreamid
}

func (p *Server) ListenFromFD() (l net.Listener, err error) {

	var fd uintptr

	if _, err = fmt.Sscan(os.Getenv("GOAGAIN_FD"), &fd); nil != err {
		return nil, err
	}

	log.Info("ListenFromFD fd :", fd)

	l, err = net.FileListener(os.NewFile(3, os.Getenv("GOAGAIN_NAME")))

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

func (p *Server) ListenFD(addr string) (net.Listener, error) {

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	return l, nil
}

func (p *Server) setEnvs() (fd uintptr, err error, strs []string) {

	v := reflect.ValueOf(p.l).Elem().FieldByName("fd").Elem()

	fd = uintptr(v.FieldByName("sysfd").Int())
	/*
		_, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFD, 0)

		if 0 != e1 {
			err = e1
			return
		}
	*/
	if err = os.Setenv("GOAGAIN_FD", fmt.Sprint(fd)); nil != err {
		return
	}
	addr := p.l.Addr()

	if err = os.Setenv(
		"GOAGAIN_NAME",
		fmt.Sprintf("%s:%s->", addr.Network(), addr.String()),
	); nil != err {
		return
	}

	strs = append(strs, fmt.Sprintf("%s=%s", "GOAGAIN_FD", fmt.Sprint(fd)))
	strs = append(strs, fmt.Sprintf("%s=%s", "GOAGAIN_NAME", fmt.Sprintf("%s:%s->", addr.Network(), addr.String())))

	return
}

func (p *Server) SetCon(con net.Conn) {

	tcp_con, ok := con.(*net.TCPConn)

	if !ok {
		log.Error("server SetCon error con not tcpcon")
		return
	}

	fild, err := tcp_con.File()

	if err != nil {
		log.Error("server SetCon file error :", err)
		return
	}

	p.files = append(p.files, fild)
}

func (p *Server) ListenAndServe() error {

	addr := p.Addr
	if addr == "" {
		addr = ":1935"
	}
	/*
		var err error
		var l net.Listener

		if os.Getenv("GRACEFUL_RESTART") == "true" {
			l, err = p.ListenFromFD()
		} else {
			log.Info("listen fd &&&&&&&&&&&&&& pid:", os.Getpid())
			l, err = p.ListenFD(addr)
		}

		if err != nil {
			return err
		}

		tcpl := l.(*net.TCPListener)

		p.l = tcpl

		//l, err := net.Listen("tcp", addr)
	*/

	addr_s, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		return err
	}

	l, err := Gracennet_net.ListenTCP("tcp", addr_s)

	if err != nil {
		return err
	}

	p.l = l

	for i := 0; i < runtime.NumCPU(); i++ {
		go p.loop(l)
	}

	return nil
}

func (srv *Server) loop(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		if srv.CloseFlag {
			log.Info("stop Accep listen pid :", os.Getpid())
			return nil
		}
		grw, e := l.Accept()
		if e != nil {
			log.Error("Accept error :", e)
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Errorf("rtmp: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		log.Info("Accept pid :", os.Getpid())
		//srv.SetCon(grw)
		tempDelay = 0
		go serve(srv, grw)
	}
}

func serve(srv *Server, con net.Conn) {
	log.Info("Accept", con.RemoteAddr(), "->", con.LocalAddr())
	con.(*net.TCPConn).SetNoDelay(true)
	conn := newconn(con, srv)
	if !handshake1(conn.buf) {
		conn.Close()
		return
	}
	log.Info("handshake", con.RemoteAddr(), "->", con.LocalAddr(), "ok")
	log.Debug("readMessage")
	msg, err := readMessage(conn)
	if err != nil {
		log.Error("NetConnecton read error", err)
		conn.Close()
		return
	}

	cmd, ok := msg.(*ConnectMessage)
	if !ok || cmd.Command != "connect" {
		log.Error("NetConnecton Received Invalid ConnectMessage ", msg)
		conn.Close()
		return
	}
	conn.app = getString(cmd.Object, "app")

	conn.objectEncoding = int(getNumber(cmd.Object, "objectEncoding"))
	log.Debug(cmd)
	log.Info(con.RemoteAddr(), "->", con.LocalAddr(), cmd, conn.app, conn.objectEncoding)
	err = sendAckWinsize(conn, 512<<10)
	if err != nil {
		log.Error("NetConnecton sendAckWinsize error", err)
		conn.Close()
		return
	}
	err = sendPeerBandwidth(conn, 512<<10)
	if err != nil {
		log.Error("NetConnecton sendPeerBandwidth error", err)
		conn.Close()
		return
	}
	err = sendStreamBegin(conn)
	if err != nil {
		log.Error("NetConnecton sendStreamBegin error", err)
		conn.Close()
		return
	}
	err = sendConnectSuccess(conn)
	if err != nil {
		log.Error("NetConnecton sendConnectSuccess error", err)
		conn.Close()
		return
	}
	conn.connected = true
	newNetStream(conn, shandler, nil, srv.Wp).readLoop()
}

func getNumber(obj interface{}, key string) float64 {
	if v, exist := obj.(Map)[key]; exist {
		return v.(float64)
	}
	return 0.0
}

func findObject(name string) (*StreamObject, bool) {
	if v, found := objects.Get(name); found {
		return v.(*StreamObject), true
	}
	return nil, false
}

func addObject(obj *StreamObject) {
	objects.Set(obj.name, obj)
}

func removeObject(name string) {
	objects.Remove(name)
}

func FindObject(name string) (*StreamObject, bool) {
	if v, found := objects.Get(name); found {
		return v.(*StreamObject), true
	}
	return nil, false
}
