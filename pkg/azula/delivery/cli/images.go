package cli

import (
	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"img", "image", "i"},
	Short:   "Manimulate with images",
	Run:     Images,
}

func init() {
	rootCmd.AddCommand(imagesCmd)
}

func Images(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cobra.CheckErr(cmd.Usage())
	}
}
