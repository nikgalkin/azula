package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	imagesCmd = &cobra.Command{
		Use:     "images",
		Aliases: []string{"img", "image", "i"},
		Short:   "Manimulate with images",
		Run:     Images,
	}
	max_entries = 0
	like        = ""
)

func init() {
	rootCmd.AddCommand(imagesCmd)
	imagesListCmd.Flags().IntVarP(&max_entries, "entries", "e", 500, "set max entries of repositories")
	imagesListCmd.Flags().StringVarP(&like, "like", "l", "", "filter images by string")
}

func Images(cmd *cobra.Command, args []string) {
	cobra.CheckErr(cmd.Usage())
	os.Exit(1)
}
