/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/advancevillage/3rd/dht"
	"github.com/spf13/cobra"
)

var (
	nodeStr string
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate info for tool",
	Long:  "Generate info for tool",

	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case len(nodeStr) > 0:
			genNode(nodeStr)
		}
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringVarP(&nodeStr, "node", "n", "", "generate nodeid for dht")
}

//ip:port,ip
func genNode(s string) {
	var nodes = strings.Split(s, ",")
	for i := range nodes {
		var (
			hostStr string
			portStr string
			err     error
		)
		if len(strings.Split(nodes[i], ":")) > 1 {
			hostStr, portStr, err = net.SplitHostPort(nodes[i])
			if err != nil {
				continue
			}
		} else {
			hostStr = nodes[i]
			portStr = "5555"
		}
		host := net.ParseIP(hostStr)
		if nil == host {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		host = host.To4()

		ipv4 := uint32(host[0]) << 24
		ipv4 |= uint32(host[1]) << 16
		ipv4 |= uint32(host[2]) << 8
		ipv4 |= uint32(host[3])

		node := dht.NewNode("udp", 0x0000, uint16(port), ipv4)
		fmt.Printf("0x%x\n", node.Encode())
	}
}
