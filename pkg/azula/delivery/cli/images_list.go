package cli

import (
	"context"

	"github.com/spf13/cobra"
)

var imagesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List images",
	Run:     ImagesList,
}

func init() {
	imagesCmd.AddCommand(imagesListCmd)
	imagesListCmd.Flags().StringP("like", "l", "", "filter images by string")
}

func ImagesList(cmd *cobra.Command, args []string) {
	ctx := context.TODO()
	like := cmd.Flag("like").Value.String()

	repos, err := meta.UC.ListReposLike(ctx, like)
	cobra.CheckErr(err)
	pickedRepos := SurveyCheckboxes("In which repositories do you want to list images?", repos)

	repoTags, err := meta.UC.GetImagesWithTags(ctx, pickedRepos)
	cobra.CheckErr(err)
	SurveyList("Found images:", repoTags)
}
