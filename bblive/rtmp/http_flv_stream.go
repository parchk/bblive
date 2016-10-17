package rtmp

import (
	"bbllive/conf"
	"bbllive/log"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
)

var (
	HEADER_BYTES = []byte{'F', 'L', 'V', 0x01, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00,
		0x12, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 11
		0x02, 0x00, 0x0a, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x61, 0x44, 0x61, 0x74, 0x61, // 13
		0x08, 0x00, 0x00, 0x00, 0x01, // 5
		0x00, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6F, 0x6E, // 10
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 9
		0x00, 0x00, 0x09, // 3
		0x00, 0x00, 0x00, 0x33}
)

var (
	HEADER_BYTES2 = []byte{'F', 'L', 'V', 0x01, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00}
)

type HttpFlvStream struct {
	notify            chan *int
	obj               *StreamObject
	streamName        string
	nsid              int
	err               error
	closed            chan bool
	lastTimestamp     uint32
	firstTimestamp    uint32
	firstTimestampSet bool
	duration          float64
	fnotify           chan *MediaFrame
	firstIframe       *MediaFrame
}

func nsid() int {
	id, _ := conf.Snow.Next()
	return int(id)
}

func NewHttpFlvStream() (s *HttpFlvStream) {
	s = new(HttpFlvStream)
	s.nsid = nsid()
	s.notify = make(chan *int, 30)
	s.fnotify = make(chan *MediaFrame, 1000)
	s.closed = make(chan bool)
	return s
}

func (s *HttpFlvStream) SetObj(o *StreamObject) {
	s.obj = o
}

func (s *HttpFlvStream) isClosed() bool {
	select {
	case _, ok := <-s.closed:
		if !ok {
			return true
		}
	default:
	}
	return false
}

func (s *HttpFlvStream) Close() error {
	if s.isClosed() {
		return nil
	}

	close(s.closed)
	close(s.notify)
	close(s.fnotify)
	return nil
}

func (s *HttpFlvStream) NsID() int {
	return s.nsid
}

func (s *HttpFlvStream) Name() string {
	return s.streamName
}

func (s *HttpFlvStream) String() string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v %v closed:%v", s.nsid, s.streamName, s.isClosed())
}

func (s *HttpFlvStream) StreamObject() *StreamObject {
	return s.obj
}

func (s *HttpFlvStream) FNotify(frame *MediaFrame) error {
	if s.isClosed() {
		return errors.New("remote connection? " + " closed")
	}
	select {
	case s.fnotify <- frame:
		return nil
	default:
		log.Debug("romode addr?" + "buffer full")
	}
	return nil
}

func (s *HttpFlvStream) Notify(idx *int) error {

	if s.isClosed() {
		return errors.New("remote connection? " + " closed")
	}

	select {
	case s.notify <- idx:
		return nil
	default:
		log.Warn("romode addr?" + "buffer full")
	}
	return nil
}

func (s *HttpFlvStream) SendHeader(w http.ResponseWriter, r *http.Request) {
	w.Write(HEADER_BYTES2)
}

func (s *HttpFlvStream) SendTag(w http.ResponseWriter, r *http.Request, data []byte, tagType byte, timestamp uint32) error {
	if timestamp < s.lastTimestamp {
		timestamp = s.lastTimestamp
	} else {
		s.lastTimestamp = timestamp
	}
	if !s.firstTimestampSet {
		s.firstTimestampSet = true
		s.firstTimestamp = timestamp
	}
	timestamp -= s.firstTimestamp
	duration := float64(timestamp) / 1000.0

	if s.duration < duration {
		s.duration = duration
	}

	headerBuf := make([]byte, 11)

	binary.BigEndian.PutUint32(headerBuf[3:7], timestamp)
	headerBuf[7] = headerBuf[3]
	binary.BigEndian.PutUint32(headerBuf[:4], uint32(len(data)))
	headerBuf[0] = tagType

	buffer := bytes.NewBuffer(headerBuf)
	buffer.Write(data)
	binary.Write(buffer, binary.BigEndian, uint32(len(data))+11)

	flvdata := buffer.Bytes()

	_, err := w.Write(flvdata)

	if err != nil {
		return err
	}

	w.(http.Flusher).Flush()

	return nil
}
func (s *HttpFlvStream) WriteLoopF(w http.ResponseWriter, r *http.Request) {

	log.Info(r.RemoteAddr, "->", r.Host, "http writeLoop running")

	defer log.Info(r.RemoteAddr, "->", r.Host, "http writeLoop stopped")

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/octet-stream")

	var (
		opened bool
		obj    *StreamObject
		err    error
		frame  *MediaFrame
	)

	s.SendHeader(w, r)
	obj = s.obj

	if obj.GetFirstVideoKeyFrame() != nil {
		key := obj.GetFirstVideoKeyFrame().Bytes()
		s.SendTag(w, r, key, RTMP_MSG_VIDEO, 0)
	} else {
		log.Error("!!!!!!! FirstVideoKeyFram is nil ")
	}

	if obj.GetFirstAudioKeyFrame() != nil {
		audio_key := obj.GetFirstAudioKeyFrame().Bytes()
		s.SendTag(w, r, audio_key, RTMP_MSG_AUDIO, 0)
		log.Error("!!!!!!! FirstAudioKeyFram is nil ")
	}

	for {
		select {
		case frame, opened = <-s.fnotify:
			if !opened {
				return
			}

			if s.firstIframe == nil {

				if frame.IFrame() {
					s.firstIframe = frame
				} else {
					continue
				}
			}

			if frame.Type == RTMP_MSG_VIDEO {

				payload := frame.Payload.Bytes()

				err = s.SendTag(w, r, payload, RTMP_MSG_VIDEO, frame.Timestamp)

			} else if frame.Type == RTMP_MSG_AUDIO {

				payload := frame.Payload.Bytes()

				err = s.SendTag(w, r, payload, RTMP_MSG_AUDIO, frame.Timestamp)
			}
		}

		if err != nil {
			break
		}

	}
}

func (s *HttpFlvStream) WriteLoop(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, "->", r.Host, "http writeLoop running")
	defer log.Info(r.RemoteAddr, "->", r.Host, "http writeLoop stopped")

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "flv-application/octet-stream")

	var (
		notify = s.notify
		opened bool
		idx    *int
		obj    *StreamObject
		gop    *MediaGop
		err    error
	)

	s.SendHeader(w, r)
	obj = s.obj

	if obj.GetFirstVideoKeyFrame() != nil {
		key := obj.GetFirstVideoKeyFrame().Bytes()
		s.SendTag(w, r, key, RTMP_MSG_VIDEO, 0)
	}

	if obj.GetFirstAudioKeyFrame() != nil {
		audio_key := obj.GetFirstAudioKeyFrame().Bytes()
		s.SendTag(w, r, audio_key, RTMP_MSG_AUDIO, 0)
	}

	for {
		select {
		case idx, opened = <-notify:
			if !opened {
				return
			}
			gop = obj.ReadGop(idx)
			if gop != nil {
				frames := gop.Frames()
				for _, frame := range frames {

					if frame.Type == RTMP_MSG_VIDEO {

						payload := frame.Payload.Bytes()

						s.SendTag(w, r, payload, RTMP_MSG_VIDEO, frame.Timestamp)

					} else if frame.Type == RTMP_MSG_AUDIO {

						payload := frame.Payload.Bytes()

						s.SendTag(w, r, payload, RTMP_MSG_AUDIO, frame.Timestamp)
					}

					if err != nil {
						break
					}

				}
			}
		}
	}
}
