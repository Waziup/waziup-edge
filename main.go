package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Waziup/waziup-edge/api"
	"github.com/Waziup/waziup-edge/mqtt"
	"github.com/Waziup/waziup-edge/tools"
	"github.com/globalsign/mgo"
)

var static http.Handler

func main() {
	// Remove date and time from logs
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	tlsCert := flag.String("crt", "", "TLS Cert File (.crt)")
	tlsKey := flag.String("key", "", "TLS Key File (.key)")

	www := flag.String("www", "/var/www", "HTTP files root")

	dbAddr := flag.String("db", "localhost:27017", "MongoDB address.")

	flag.Parse()

	////////////////////

	log.Println("Waziup API Server")
	log.Println("--------------------")

	////////////////////

	if *www != "" {
		log.Printf("[WWW  ] Serving from %q\n", *www)
		static = http.FileServer(http.Dir(*www))
	} else {
		log.Printf("[WWW  ] Not serving www files.\n")
	}

	////////////////////

	log.Printf("[DB   ] Dialing MongoDB at %q...\n", *dbAddr)

	db, err := mgo.Dial("mongodb://" + *dbAddr + "/?connect=direct")
	if err != nil {
		log.Println("[DB   ] MongoDB client error:\n", err)
	} else {

		api.DBSensorValues = db.DB("waziup").C("sensor_values")
		api.DBActuatorValues = db.DB("waziup").C("actuator_values")
		api.DBDevices = db.DB("waziup").C("devices")
	}

	////////////////////

	log.Printf("[     ] Local device id is %q.\n", api.GetLocalID())

	////////////////////

	if *tlsCert != "" && *tlsKey != "" {

		cert, err := ioutil.ReadFile(*tlsCert)
		if err != nil {
			log.Println("Error reading", *tlsCert)
			log.Fatalln(err)
		}

		key, err := ioutil.ReadFile(*tlsKey)
		if err != nil {
			log.Println("Error reading", *tlsKey)
			log.Fatalln(err)
		}

		pair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			log.Println("TLS/SSL 'X509KeyPair' Error")
			log.Fatalln(err)
		}

		cfg := &tls.Config{Certificates: []tls.Certificate{pair}}

		ListenAndServeHTTPS(cfg)
		ListenAndServeMQTTTLS(cfg)
	}

	ListenAndServerMQTT()
	ListenAndServeHTTP() // will block
}

///////////////////////////////////////////////////////////////////////////////

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func (resp *ResponseWriter) WriteHeader(statusCode int) {
	resp.status = statusCode
	resp.ResponseWriter.WriteHeader(statusCode)
}

////////////////////

func Serve(resp http.ResponseWriter, req *http.Request) {
	wrapper := ResponseWriter{resp, 200}

	if static != nil {
		if strings.HasPrefix(req.RequestURI, "/www/") {
			req.RequestURI = req.RequestURI[4:]
			req.URL.Path = req.URL.Path[4:]
			static.ServeHTTP(&wrapper, req)

			log.Printf("[WWW  ] (%s) %d %s \"/www%s\"\n",
				req.RemoteAddr,
				wrapper.status,
				req.Method,
				req.RequestURI)
			return
		}
	}

	if req.Method == http.MethodPut || req.Method == http.MethodPost {

		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(resp, "400 Bad Request", http.StatusBadRequest)
			return
		}
		req.Body = &tools.ClosingBuffer{bytes.NewBuffer(body)}
	}

	if req.Method == "PUBLISH" {
		req.Method = http.MethodPost
		router.ServeHTTP(&wrapper, req)
		req.Method = "PUBLISH"
	} else {
		router.ServeHTTP(&wrapper, req)
	}

	log.Printf("[%s] %s %d %s %q\n",
		req.Header.Get("X-Tag"),
		req.RemoteAddr,
		wrapper.status,
		req.Method,
		req.RequestURI)

	if cbuf, ok := req.Body.(*tools.ClosingBuffer); ok {
		// log.Printf("[DEBUG] Body: %s\n", cbuf.Bytes())
		msg := &mqtt.Message{
			QoS:   0,
			Topic: req.RequestURI[1:],
			Data:  cbuf.Bytes(),
		}

		// if wrapper.status >= 200 && wrapper.status < 300 {
		if req.Method == http.MethodPut || req.Method == http.MethodPost {
			mqttServer.Publish(nil, msg)
		}
		// }
	}

}