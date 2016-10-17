package rtmp

import (
	"bbllive/log"
	"fmt"
	"net/http"
)

type HttpFlvPlayHandle struct {
}

func (*HttpFlvPlayHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println("request ParseForm error")
		http.NotFound(w, r)
		return
	}

	id := r.Form.Get("s")

	objname := fmt.Sprintf("live/%s", id)

	obj, found := FindObject(objname)

	if !found {
		log.Error("object not find key :", objname)
		http.NotFound(w, r)
		return
	}

	stream := NewHttpFlvStream()
	stream.SetObj(obj)

	obj.HttpAttach(stream)

	//stream.WriteLoop(w, r)
	stream.WriteLoopF(w, r)
	stream.Close()

}

func ListenAndServerHttpFlv(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/play", &HttpFlvPlayHandle{})
	server := &http.Server{Addr: addr, Handler: mux}
	return server.ListenAndServe()
}
