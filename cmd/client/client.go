package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	client2 "github.com/fabric-creed/fabric-hub/pkg/client"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"time"
)

func main() {
	client, err := client2.NewHubClient("orderer1.example.com:1000", client2.ClientConfig{
		UseTLS:               true,
		ClientKeyPath:        "",
		ClientCertPath:       "",
		ClientRootCACertPath: "",
		ServerRootCAPath:     "./test/ca.crt",
		IsGm:                 true,
	})
	if err != nil {
		panic(err)
	}
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	payload := pb.FabricPayload{
		ChannelID:     "mychannel",
		ChainCodeName: "fabcar",
		FncName:       "QueryCar",
		Args:          []string{"CAR1"},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	resp, err := client.NoTransactionCall(
		"1411931388202418176",
		"1411931388202418176",
		node.Generate().String(),
		nil,
		data,
		time.Now().Unix(),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Payload)
	r, err := fabric.DecodeInvokeChainCodeResponse(resp.Payload)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r.PayloadData))
}
