package main

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"net"
)

func main() {
	command := &cobra.Command{}
	command.AddCommand(GenerateID())
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}

func GenerateID() *cobra.Command {
	generateMigrateCommand := &cobra.Command{
		Use:   "id",
		Short: "use to generate channel id",
		Run: func(cmd *cobra.Command, args []string) {
			nodeID, err := lower16BitPrivateIP()
			if err != nil {
				panic(errors.Wrap(err, "failed to get node id"))
			}
			node, err := snowflake.NewNode(nodeID)
			if err != nil {
				panic(errors.Wrap(err, "failed to generate id"))
			}
			fmt.Printf("chanel id: %d \n", node.Generate().Int64())
		},
	}

	return generateMigrateCommand
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (int64, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return (int64(ip[2])<<8 + int64(ip[3])) % 1024, nil
}
