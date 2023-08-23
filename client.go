package vcs

import (
	"errors"

	gitclient "github.com/cidverse/go-vcs/git"
	"github.com/cidverse/go-vcs/vcsapi"
)

var MockClient vcsapi.Client

func GetVCSClient(dir string) (vcsapi.Client, error) {
	// mocked client
	if MockClient != nil {
		return MockClient, nil
	}

	// git
	cg, _ := gitclient.NewGitClient(dir)
	if cg.Check() {
		return cg, nil
	}

	return nil, errors.New("directory is not a vcs repository")
}

func GetVCSClientCloneRemote(cloneUrl string, dir string, branch string) (vcsapi.Client, error) {
	// mocked client
	if MockClient != nil {
		return MockClient, nil
	}

	// git
	cg, _ := gitclient.NewGitClientCloneFromURL(cloneUrl, dir, branch)
	if cg.Check() {
		return cg, nil
	}

	return nil, errors.New("directory is not a vcs repository")
}
