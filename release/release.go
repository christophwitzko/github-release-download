package release

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
)

type Asset struct {
	Version  string
	OS       string
	Arch     string
	FileName string
	URL      string
}

type GithubClient struct {
	Client *github.Client
	lock   chan struct{}
}

func NewClient(token string) *GithubClient {
	return &GithubClient{github.NewClient(oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))), make(chan struct{}, 1)}
}

func (c *GithubClient) GetLatestReleaseAssets(ctx context.Context, owner, repo string) ([]*Asset, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c.lock <- struct{}{}:
	}
	release, _, err := c.Client.Repositories.GetLatestRelease(ctx, owner, repo)
	<-c.lock
	if err != nil {
		return nil, err
	}
	ret := make([]*Asset, len(release.Assets))
	for i, ra := range release.Assets {
		ret[i] = &Asset{
			Version:  release.GetTagName(),
			FileName: ra.GetName(),
			URL:      ra.GetBrowserDownloadURL(),
		}
		osarch := strings.Split(ret[i].FileName, "_")
		if len(osarch) < 2 {
			continue
		}
		ret[i].OS = strings.ToLower(osarch[len(osarch)-2])
		arch := strings.SplitN(strings.ToLower(osarch[len(osarch)-1]), ".", 2)
		ret[i].Arch = arch[0]
	}
	return ret, nil
}

func (c *GithubClient) GetLatestDownloadUrl(ctx context.Context, owner, repo, os, arch string) (string, error) {
	assets, err := c.GetLatestReleaseAssets(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	os = strings.ToLower(os)
	arch = strings.ToLower(arch)
	for _, asset := range assets {
		if asset.OS == os && asset.Arch == arch {
			return asset.URL, nil
		}
	}
	return "", nil
}
