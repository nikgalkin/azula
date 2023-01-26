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
}

func ImagesList(cmd *cobra.Command, args []string) {
	ctx := context.TODO()

	repos, err := meta.UC.ListReposLike(ctx, like, max_entries)
	cobra.CheckErr(err)
BACK:
	pickedRepos := SurveyList("In which repositories do you want to list images?", repos)

	repoTags, err := meta.UC.GetImagesWithTags(ctx, []string{pickedRepos})
	cobra.CheckErr(err)
	back := SurveyList("Found images:", append(repoTags, mgmtBack))
	if back == mgmtBack {
		goto BACK
	}
}
