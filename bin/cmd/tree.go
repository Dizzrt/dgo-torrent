package cmd

import (
	"fmt"
	"os"
	"strings"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
	"github.com/Dizzrt/dgo-torrent/dlog"
	"github.com/spf13/cobra"
)

// treeCmd represents the tree command
var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "list all files",
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

		if tf.Info.IsMutiFile {
			printTree(tf.Info.MutiFiles, make([]string, 0), 0)
		} else {
			fmt.Printf("%s [%s]\n", tf.Info.Name, formatSize(tf.Info.Length))
		}

		return nil
	},
}

func formatSize(size int64) string {
	s := float64(size)

	index := 0
	units := []string{"B", "KB", "MB", "GB"}
	for ; index < 4; index++ {
		if s < 1024 {
			break
		}

		s = s / 1024
	}

	return fmt.Sprintf("%.2f %s", s, units[index])
}

func printTree(tmf dgotorrent.TorrentMutiFile, prefixs []string, level int) {
	if level == 0 {
		fmt.Printf(". %s\n", tmf.Name)
	}

	l := len(tmf.SubsOrder)
	lprefix := strings.Join(prefixs, "")
	for i, subName := range tmf.SubsOrder {
		fmt.Print(lprefix)

		var sprefix string
		if i == l-1 {
			sprefix = "└── "
		} else {
			sprefix = "├── "
		}
		fmt.Print(sprefix)

		sub := tmf.Subs[subName]
		if sub.Type == dgotorrent.TMF_TYPE_DIRECTORY {
			fmt.Printf("\033[1;96;40m%s\033[0m\n", sub.Name)
		} else if sub.Type == dgotorrent.TMF_TYPE_FILE {

			fmt.Printf("%s [%s]\n", sub.Name, formatSize(sub.Length))
		} else {
			dlog.Error(fmt.Errorf(""))
		}

		if sub.Type == dgotorrent.TMF_TYPE_DIRECTORY {
			printTree(sub, append(prefixs, "│\u00A0\u00A0 "), level+1)
		}
	}
}

func init() {
	rootCmd.AddCommand(treeCmd)
}
