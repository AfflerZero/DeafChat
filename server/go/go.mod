module github.com/AfflerZero/DeafChat

go 1.26.0

require (
	github.com/gorilla/websocket v1.5.3
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.10.9
	github.com/wneessen/go-mail v0.6.2
	golang.org/x/crypto v0.56.0
	golang.org/x/time v0.12.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	golang.org/x/tools/go/expect v0.1.1-deprecated // indirect
	honnef.co/go/tools v0.6.0 // indirect
)

tool honnef.co/go/tools/cmd/staticcheck
