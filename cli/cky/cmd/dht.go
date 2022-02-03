/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/advancevillage/3rd/dht"
	"github.com/advancevillage/3rd/logx"
	"github.com/spf13/cobra"
)

var (
	nodeHex uint64
	seedStr string
)

// dhtCmd represents the dht command
var dhtCmd = &cobra.Command{
	Use:   "dht",
	Short: "Manage dht p2p system",
	Long:  "Manage dht p2p system",

	Run: func(cmd *cobra.Command, args []string) {
		dhtSrv(nodeHex)
	},
}

func init() {
	rootCmd.AddCommand(dhtCmd)

	//节点16进制
	dhtCmd.Flags().Uint64VarP(&nodeHex, "node", "n", 0x000000000000000, "nodeId display in hex")

	//种子节点
	dhtCmd.Flags().StringVarP(&seedStr, "seed", "s", "", "seed for dht")

	//必须传
	dhtCmd.MarkFlagRequired("node")
}

func dhtSrv(node uint64) {
	//解析参数
	enc := dht.Decode(node)

	seeds := strings.Split(seedStr, ",")
	seed := []uint64{}
	for _, v := range seeds {
		v = strings.Replace(v, "0x", "", -1)
		vv, err := strconv.ParseUint(v, 16, 64)
		if err != nil {
			continue
		}
		seed = append(seed, vv)
	}

	var ctx, cancel = context.WithCancel(context.TODO())
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
		}
		signal.Stop(c)
		close(c)
	}()

	var logger, err = logx.NewLogger("info")

	if err != nil {
		return
	}
	port := enc.Port()
	ipv4 := enc.Ipv4()

	addr := fmt.Sprintf("%d.%d.%d.%d:%d", byte(ipv4>>24), byte(ipv4>>16), byte(ipv4>>8), byte(ipv4), port)

	logger.Infow(ctx, "dht seed list", "seeds", seed)

	srv, err := dht.NewDHT(ctx, logger, &dht.DHTCfg{
		Fix:     5,
		Refresh: 2,
		Evolut:  10,
		Network: "udp",
		Zone:    enc.Zone(),
		Addr:    addr,
		Seeds:   seed,
		Alpha:   3,
		K:       4,
	})
	if err != nil {
		return
	}
	go srv.Start()

	for {
		select {
		case <-ctx.Done():
		default:
			time.Sleep(time.Second * 2)
			logger.Infow(ctx, "dht show", "info", srv.Show(ctx))
		}
	}
}
