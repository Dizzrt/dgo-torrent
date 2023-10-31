package cmdbencode

import (
	"github.com/spf13/cobra"
)

// flags
var (
	inputFile string
)

var BencodeCmd = &cobra.Command{
	Use:   "bencode",
	Short: "Bencode Utility",
}

func init() {
	BencodeCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "target file path")
}
