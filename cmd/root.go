package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tasker",
	Short: "TFS task creator",
	Long:  `This application is a tool to rapidly create TFS tasks and synchronize them with wiki.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("baseAddress", "http://msk-tfs-t.infotecs-nt:8080/tfs/SrvNccCollection")
	viper.SetDefault("project", "NSMS")
	viper.SetDefault("team", "SMP")
	viper.SetDefault("discipline", "Development")
	viper.SetDefault("user", "ANON")
	viper.SetDefault("personalAccessToken", "")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tasker")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		_ = viper.SafeWriteConfig()
	}
}
