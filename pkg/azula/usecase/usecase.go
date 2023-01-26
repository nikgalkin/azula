package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/nikgalkin/azula/pkg/azula/repository/docker"

	"github.com/docker/distribution"
)

type usecase struct {
	Registry docker.Manager
}

type ManUsecase interface {
	ListReposLike(context.Context, string, int) ([]string, error)
	GetImagesWithTags(context.Context, []string) ([]string, error)
	DeleteImageByTag(context.Context, []string) error
}

func New(reg docker.Manager) ManUsecase {
	return &usecase{
		Registry: reg,
	}
}

func (u *usecase) ListReposLike(ctx context.Context, like string, max_entries int) ([]string, error) {
	return u.Registry.ListReposLike(ctx, like, max_entries)
}

func (u *usecase) GetImagesWithTags(ctx context.Context, repos []string) ([]string, error) {
	res := make([]string, 0, 4)
	for _, repo := range repos {
		r, err := u.Registry.GetRepo(ctx, repo)
		if err != nil {
			return []string{}, err
		}
		tags, err := r.Tags(ctx).All(ctx)
		if err != nil {
			return []string{}, err
		}
		joinRepoWithTag(repo, tags)
		res = append(res, tags...)
	}
	return res, nil
}

func (u *usecase) DeleteImageByTag(ctx context.Context, repoTags []string) error {
	for _, v := range repoTags {
		rt := strings.Split(v, ":")
		repo := rt[0]
		tag := rt[1]
		if len(repo) < 1 || len(tag) < 1 {
			return fmt.Errorf("repo or tag empty. repo: '%s', tag: '%s'", repo, tag)
		}
		r, err := u.Registry.GetRepo(ctx, rt[0])
		if err != nil {
			return err
		}
		m, err := r.Manifests(ctx, distribution.WithTag(tag))
		if err != nil {
			return err
		}
		desc, err := u.Registry.GetV2Descriptor(ctx, repo, tag)
		if err != nil {
			return err
		}
		err = m.Delete(ctx, desc.Digest)
		if err != nil {
			return err
		}
	}
	return nil
}

func joinRepoWithTag(repo string, tags []string) {
	for i, t := range tags {
		tags[i] = repo + ":" + t
	}
}
