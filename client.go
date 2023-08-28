package vcs

import (
	"errors"

	gitclient "github.com/cidverse/go-vcs/git"
	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/go-git/go-git/v5/plumbing/transport"
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

func GetVCSClientCloneRemote(cloneUrl string, dir string, branch string, auth transport.AuthMethod) (vcsapi.Client, error) {
	// mocked client
	if MockClient != nil {
		return MockClient, nil
	}

	// git
	cg, _ := gitclient.NewGitClientCloneFromURL(cloneUrl, dir, branch, auth)
	if cg.Check() {
		return cg, nil
	}

	return nil, errors.New("directory is not a vcs repository")
}
