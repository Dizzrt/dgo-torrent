package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
	"github.com/spf13/cobra"
)

type MutiFileType uint8

const (
	MFT_DIRECTORY = iota
	MFT_FILE
)

type MutiFile struct {
	Type      MutiFileType
	Name      string
	Size      int
	Subs      map[string]MutiFile
	SubsOrder []string
}

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

		if !tf.Info.IsMutiFile {
			fmt.Println(tf.Info.Name)
		} else {
			// jb, _ := json.Marshal(tf.Info.MutiFiles)
			// fmt.Println(string(jb))
			p(tf.Info.MutiFiles)
		}

		return nil
	},
}

func buildDir(upper MutiFile, path []string) MutiFile {
	if len(path) == 1 {
		upper.Subs[path[0]] = MutiFile{
			Type:      MFT_FILE,
			Name:      path[0],
			Size:      0,
			Subs:      nil,
			SubsOrder: nil,
		}

		upper.SubsOrder = append(upper.SubsOrder, path[0])
		return upper
	}

	var m MutiFile
	if temp, ok := upper.Subs[path[0]]; ok {
		m = temp
	} else {
		m = MutiFile{
			Type:      MFT_DIRECTORY,
			Name:      path[0],
			Size:      0,
			Subs:      make(map[string]MutiFile),
			SubsOrder: make([]string, 0),
		}
	}
	m = buildDir(m, path[1:])

	upper.Subs[path[0]] = m
	upper.SubsOrder = append(upper.SubsOrder, path[0])

	return upper
}

func p(f []dgotorrent.TorrentMutiFile) {
	m := MutiFile{
		Type:      MFT_DIRECTORY,
		Name:      "root",
		Size:      0,
		Subs:      make(map[string]MutiFile),
		SubsOrder: make([]string, 0),
	}

	for _, v := range f {
		m = buildDir(m, v.Path)
		fmt.Println(v.Path)
		fmt.Println()
	}

	j, _ := json.Marshal(m)
	fmt.Println(string(j))
}

func init() {
	rootCmd.AddCommand(treeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// treeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// treeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
