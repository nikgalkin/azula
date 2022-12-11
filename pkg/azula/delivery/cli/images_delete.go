package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var imagesDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d", "del"},
	Short:   "Delete images",
	Run:     ImagesDelete,
}

func init() {
	imagesCmd.AddCommand(imagesDeleteCmd)
	imagesDeleteCmd.Flags().StringP("like", "l", "", "filter images by string")
}

func ImagesDelete(cmd *cobra.Command, args []string) {
	ctx := context.TODO()
	like := cmd.Flag("like").Value.String()

	repos, err := meta.UC.ListReposLike(ctx, like)
	cobra.CheckErr(err)
	pickedRepos := SurveyCheckboxes("In which repositories do you want to delete images?", repos)

	repoTags, err := meta.UC.GetImagesWithTags(ctx, pickedRepos)
	cobra.CheckErr(err)
	pickedTags := SurveyCheckboxes("Which tags do you want to delete?", repoTags)

	if len(pickedTags) < 1 {
		fmt.Println("Tags for images", strings.Join(pickedRepos, ", "), "not found")
		return
	}
	err = meta.UC.DeleteImageByTag(ctx, pickedTags)
	cobra.CheckErr(err)
	fmt.Println("Deleted next images:", strings.Join(pickedTags, ", "))
}
