package api

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/Waziup/wazigate-edge/tools"
	"github.com/globalsign/mgo"
)

// GetLocalID returns the ID of this device
func GetLocalID() string {
	interfs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, interf := range interfs {
		addr := interf.HardwareAddr.String()
		if addr != "" {
			return strings.ReplaceAll(addr, ":", "")
		}
	}
	return ""
}

func serveIter(w io.Writer, iter *mgo.Iter, interf interface{}) {
	w.Write([]byte{'['})
	first := true
	for iter.Next(interf) {
		if first == false {
			w.Write([]byte{','})
		}
		first = false
		data, _ := json.Marshal(interf)
		w.Write(data)
	}
	w.Write([]byte{']'})
	iter.Close()
}

func unmarshalRequestBody(req *http.Request, i interface{}) error {
	body, err := tools.ReadAll(req.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, i)
	if err != nil {
		return err
	}
	return nil
}
