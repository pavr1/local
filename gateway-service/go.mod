module gateway-service

go 1.23

require (
	github.com/gorilla/mux v1.8.1
	github.com/sirupsen/logrus v1.9.3
	shared v0.0.0
)

replace shared => ../shared

require (
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)
