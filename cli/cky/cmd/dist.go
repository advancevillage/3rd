/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/advancevillage/3rd/dht"
	"github.com/spf13/cobra"
)

// distCmd represents the dist command
var distCmd = &cobra.Command{
	Use:   "dist",
	Short: "Dist of src and dst",
	Long:  "Dist of src and dst",
	Run: func(cmd *cobra.Command, args []string) {
		var src = dht.Decode(srcU64)
		var dst = dht.Decode(dstU64)
		fmt.Println(dht.XOR(src, dst))
	},
}

var (
	srcU64 uint64
	dstU64 uint64
)

func init() {
	rootCmd.AddCommand(distCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// distCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// distCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	distCmd.Flags().Uint64VarP(&srcU64, "src", "s", 0x000000000000000, "src display in hex")
	distCmd.Flags().Uint64VarP(&dstU64, "dst", "d", 0x000000000000000, "dst display in hex")
	//必须传
	distCmd.MarkFlagRequired("src")
	distCmd.MarkFlagRequired("dst")
}
