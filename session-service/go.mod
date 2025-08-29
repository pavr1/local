module session-service

go 1.23

toolchain go1.24.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/crypto v0.17.0
	shared v0.0.0
)

require golang.org/x/sys v0.15.0 // indirect

replace shared => ../shared
