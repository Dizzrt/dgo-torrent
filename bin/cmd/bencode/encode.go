package cmdbencode

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Args:  cobra.MaximumNArgs(1),
	Short: "Read an integer, string, list, or dictionary from parameter or a file and encode it into a bencoded string.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputFile == "" && len(args) < 1 {
			return fmt.Errorf("param error, requires a string or file path marked with flag '-i'")
		}

		var r io.Reader
		if inputFile != "" {
			file, err := os.Open(inputFile)
			if err != nil {
				return err
			}
			defer file.Close()

			r = file
		} else if len(args) > 0 {
			str := args[0]
			r = strings.NewReader(str)
		}

		return encode(r)
	},
}

func encode(r io.Reader) error {
	// TODO

	return nil
}

func init() {
	BencodeCmd.AddCommand(encodeCmd)
}
