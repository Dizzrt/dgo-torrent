package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
	"github.com/spf13/cobra"
)

// flags
var (
	withPieces bool
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Display information about the contents of a torrent file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("param error, requires a torrent file")
		}

		path := args[0]
		fileInfo, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("target file does not exist")
			} else {
				return err
			}
		}

		if !fileInfo.Mode().IsRegular() {
			return fmt.Errorf("the target file is not a valid file")
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		tf, err := dgotorrent.NewTorrentFile(file)
		if err != nil {
			return err
		}

		if !withPieces {
			tf.Info.PieceHashes = make([][20]byte, 0)
		}

		res, err := json.Marshal(tf)
		if err != nil {
			return err
		}

		fmt.Println(string(res))
		if !withPieces {
			fmt.Println("\nThe pieces has been hidden because it is too long, please use flag -p or --pieces if you don't want to hide it.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.Flags().BoolVarP(&withPieces, "pieces", "p", false, "read the hash list which is a concatenation of each piece's SHA-1 hash")
}
