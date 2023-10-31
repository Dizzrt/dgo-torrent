package cmd

import (
	"os"

	cmdbencode "github.com/Dizzrt/dgo-torrent/bin/cmd/bencode"
	"github.com/spf13/cobra"
)

var (
	// cfgFile string

	// flags
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dgo-torrent",
	Short: "This is a free and open source torrent downloader",
	Long: `
	This application is a free and open source torrent downloader that follows the MIT protocol.

	source: github.com/Dizzrt/dgo-torrent`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cmdbencode.BencodeCmd)

	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bin.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "use verbose output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Find home directory.
// 		home, err := os.UserHomeDir()
// 		cobra.CheckErr(err)

// 		// Search config in home directory with name ".bin" (without extension).
// 		viper.AddConfigPath(home)
// 		viper.SetConfigType("yaml")
// 		viper.SetConfigName(".bin")
// 	}

// 	viper.AutomaticEnv() // read in environment variables that match

// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
// 	}
// }
