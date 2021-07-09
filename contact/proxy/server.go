package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"log"
)

func main() {
	if err := shim.Start(new(Proxy)); err != nil {
		log.Printf("failed to start contract: %v", err)
	}
}
