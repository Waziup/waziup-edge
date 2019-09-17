module github.com/Waziup/wazigate-edge

go 1.12

require (
	github.com/Waziup/wazigate-edge/mqtt v0.0.0-00010101000000-000000000000
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gorilla/websocket v1.4.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kr/pretty v0.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/Waziup/wazigate-edge/mqtt => ./mqtt
