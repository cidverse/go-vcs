package gitclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSignature = &object.Signature{Name: "Test", Email: "test@example.com", When: time.Unix(1700000000, 0)}

// setupTagTestRepo creates a repository with 5 commits, an annotated tag v1.0.0 on commit 2 and a lightweight tag v1.1.0 on commit 4
func setupTagTestRepo(t *testing.T) (dir string, commits []plumbing.Hash) {
	t.Helper()
	dir = t.TempDir()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte{byte('a' + i)}, 0o644))
		_, err = wt.Add("file.txt")
		require.NoError(t, err)
		hash, commitErr := wt.Commit("feat: commit "+string(rune('a'+i)), &git.CommitOptions{Author: testSignature, Committer: testSignature})
		require.NoError(t, commitErr)
		commits = append(commits, hash)
	}

	_, err = repo.CreateTag("v1.0.0", commits[1], &git.CreateTagOptions{Tagger: testSignature, Message: "release v1.0.0"})
	require.NoError(t, err)
	_, err = repo.CreateTag("v1.1.0", commits[3], nil)
	require.NoError(t, err)

	return dir, commits
}

func TestFindCommitsBetweenAnnotatedTag(t *testing.T) {
	dir, commits := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	result, err := client.FindCommitsBetween(nil, &vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"}, false, 0)
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, commits[4].String(), result[0].Hash)
	assert.Equal(t, commits[2].String(), result[2].Hash)
}

func TestFindCommitsBetweenLightweightTag(t *testing.T) {
	dir, commits := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	result, err := client.FindCommitsBetween(nil, &vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"}, false, 0)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, commits[4].String(), result[0].Hash)
}

func TestFindCommitsBetweenFromAnnotatedTag(t *testing.T) {
	dir, commits := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	result, err := client.FindCommitsBetween(&vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"}, nil, false, 0)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, commits[1].String(), result[0].Hash)
	assert.Equal(t, []vcsapi.VCSRef{{Type: "tag", Value: "v1.0.0", Hash: commits[1].String()}}, result[0].Tags)
}

func TestFindCommitsBetweenUnreachableBoundary(t *testing.T) {
	dir, commits := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	// v1.1.0 is not an ancestor of v1.0.0
	_, err = client.FindCommitsBetween(&vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"}, &vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"}, false, 0)
	assert.Error(t, err)

	// limit is reached before the boundary
	result, err := client.FindCommitsBetween(nil, &vcsapi.VCSRef{Hash: commits[0].String()}, false, 2)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFindCommitsBetweenInvalidRefs(t *testing.T) {
	dir, _ := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	_, err = client.FindCommitsBetween(&vcsapi.VCSRef{Type: "tag", Value: "v9.9.9"}, nil, false, 0)
	assert.Error(t, err)

	_, err = client.FindCommitsBetween(nil, &vcsapi.VCSRef{Type: "tag", Value: "v9.9.9"}, false, 0)
	assert.Error(t, err)
}

func TestGetTagsResolvesAnnotatedTags(t *testing.T) {
	dir, commits := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	assert.ElementsMatch(t, []vcsapi.VCSRef{
		{Type: "tag", Value: "v1.0.0", Hash: commits[1].String()},
		{Type: "tag", Value: "v1.1.0", Hash: commits[3].String()},
	}, client.GetTags())
	assert.Equal(t, []vcsapi.VCSRef{{Type: "tag", Value: "v1.0.0", Hash: commits[1].String()}}, client.GetTagsByHash(commits[1].String()))

	release, err := client.FindLatestRelease(true)
	require.NoError(t, err)
	assert.Equal(t, commits[3].String(), release.Hash)
}

func TestDiffAnnotatedTag(t *testing.T) {
	dir, _ := setupTagTestRepo(t)
	client, err := NewGitClient(dir)
	require.NoError(t, err)

	diff, err := client.Diff(&vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"}, &vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"})
	require.NoError(t, err)
	assert.Len(t, diff, 1)
}
