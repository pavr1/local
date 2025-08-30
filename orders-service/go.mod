module orders-service

go 1.23

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/sirupsen/logrus v1.9.3
	shared v0.0.0
)

require (
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace shared => ../shared
