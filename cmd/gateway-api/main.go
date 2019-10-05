package main

import (
	"log"
	"github.com/DiGregory/tfs-auction/internal/lot"
	"github.com/DiGregory/tfs-auction/internal/user"
)

type gateWayApp struct {
	Addr        string
	LotStorage  lot.LotsStorage
	UserStorage user.UsersStorage
}

func CreateGateWayApp(addr string, lotStorageDSN string, userStorageDSN string) (*gateWayApp, error) {
	lotStorage, err := lot.StorageConnect(lotStorageDSN)
	if err != nil {
		return nil, err
	}
	usersStorage, err := user.StorageConnect(userStorageDSN)
	if err != nil {
		return nil, err
	}
	return &gateWayApp{Addr: addr, LotStorage: *lotStorage, UserStorage: *usersStorage}, nil
}

func main() {
	DSN := "user=postgres password=1234 dbname=test sslmode=disable"
	myApp, err := CreateGateWayApp(":5000", DSN, DSN)
	if err != nil {
		log.Fatal("can`t run gateway app")
	}
	err = myApp.CreateGatewayHandler()
	if err != nil {
		log.Fatal("can`t run gateway app")
	}
}
