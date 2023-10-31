package cmdbencode

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Dizzrt/dgo-torrent/bencode"
	"github.com/spf13/cobra"
)

// flags
var (
	outputJson bool
)

var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode a bencoded string or file into its corresponding data format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputFile == "" && len(args) < 1 {
			return fmt.Errorf("param error, requires a string or a file path marked with flag '-i'")
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

		return decode(r)
	},
}

func decode(r io.Reader) error {
	res, err := bencode.Unmarshal(r)
	if err != nil {
		return err
	}

	if !outputJson {
		fmt.Println(res)
	} else {
		j, err := json.Marshal(res)
		if err != nil {
			fmt.Println(res)
		} else {
			fmt.Println(string(j))
		}
	}
	return nil
}

func init() {
	BencodeCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().BoolVarP(&outputJson, "json", "j", false, "if possible, output the decoding results in JSON format.")
}
