package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	debugFlag  = "debug"
	configFlag = "config"
)

var (
	cfgFile *string
	rootCmd = &cobra.Command{
		Aliases: []string{},
		Short:   shortDescription,
		Long:    longDescription,
	}
)

// Execute executes the root command.
func Execute(module, tag string) error {
	rootCmd.Version = fmt.Sprintf("%s@%s", module, tag)
	parts := strings.Split(module, "/")
	name := parts[len(parts)-1]
	rootCmd.Use = name + " [flags] [command]"
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	cfgFile = rootCmd.PersistentFlags().StringP(configFlag, "c", "", "use this configuration file")
	addBoolFlag(rootCmd.PersistentFlags(), debugFlag, false, "enable debug messages")
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	if *cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(*cfgFile)
	} else {
		// Find home directory.
		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/app")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "problem reading %s: %s\n", viper.ConfigFileUsed(), err)
	} else {
		fmt.Fprintln(os.Stderr, "using config file:", viper.ConfigFileUsed())
	}
}

func addBoolFlag(
	flags *pflag.FlagSet,
	name string,
	defaultValue bool,
	help string,
) {
	flags.Bool(name, defaultValue, help)
	if err := viper.BindPFlag(name, flags.Lookup(name)); err != nil {
		cobra.CheckErr(err)
	}
}

func addStringFlag(flags *pflag.FlagSet, name string, defaultValue string, help string) {
	flags.String(name, defaultValue, help)
	if err := viper.BindPFlag(name, flags.Lookup(name)); err != nil {
		cobra.CheckErr(err)
	}
}

var _ = addStringArrayFlag

func addStringArrayFlag(flags *pflag.FlagSet, name string, defaultValue []string, help string) {
	flags.StringArray(name, defaultValue, help)
	if err := viper.BindPFlag(name, flags.Lookup(name)); err != nil {
		cobra.CheckErr(err)
	}
}
