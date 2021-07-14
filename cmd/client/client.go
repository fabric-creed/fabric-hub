package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	client2 "github.com/fabric-creed/fabric-hub/pkg/client"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"time"
)

func main() {
	grpc, err := client2.NewGRPCClient(
		"",
		"",
		"",
		"./test/ca.crt",
		true,
	)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewConnection(fmt.Sprintf("%s:%d", "orderer1.example.com", 1000))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	payload := pb.FabricPayloadRequest{
		ChannelName:   "mychannel",
		ChainCodeName: "fabcar",
		FncName:       "QueryCar",
		Args:          []string{"CAR1"},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	resp, err := pb.NewHubClient(conn).NoTransactionCall(context.Background(), &pb.NoTransactionCallRequest{
		From:          "1411931467332157440",
		To:            "1411931388202418176",
		TransactionID: node.Generate().String(),
		StepID:        "1",
		Payload:       data,
		Timestamp:     time.Now().Unix(),
	})
	if err != nil {
		panic(err)
	}
	r, err := fabric.DecodeInvokeChainCodeResponse(resp.Payload)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r.PayloadData))
}
