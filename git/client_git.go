package gitclient

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/rs/zerolog/log"
)

const REF_TAGS = "refs/tags/"
const REF_HEADS = "refs/heads/"
const REF_REMOTES = "refs/remotes/"
const TAGS = "tags/"

type GitClient struct {
	dir        string
	repo       *git.Repository
	isShallow  bool
	tags       []vcsapi.VCSRef
	tagsByHash map[string][]vcsapi.VCSRef
}

func NewGitClient(dir string) (vcsapi.Client, error) {
	c := GitClient{dir: dir}
	if !c.Check() {
		return nil, errors.New("is not a git repository")
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, errors.New("failed to open git repository at " + dir + ": " + err.Error())
	}
	c.repo = repo
	c.isShallow = fileExists(filepath.Join(dir, ".git", "shallow"))

	return c, nil
}

func NewGitClientCloneFromURL(cloneURL string, localDir string, branch string, auth transport.AuthMethod) (vcsapi.Client, error) {
	// clone repository (shallow)
	repo, err := git.PlainClone(localDir, false, &git.CloneOptions{
		URL:           cloneURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
		SingleBranch:  true,
		Auth:          auth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository from %s to %s: %w", cloneURL, localDir, err)
	}

	// open
	c := GitClient{
		dir:  localDir,
		repo: repo,
	}
	if !c.Check() {
		return nil, errors.New("is not a git repository")
	}
	c.isShallow = fileExists(filepath.Join(localDir, ".git", "shallow"))

	return c, nil
}

func (c GitClient) Check() bool {
	if _, err := os.Stat(path.Join(c.dir, ".git")); !os.IsNotExist(err) {
		return true
	}

	return false
}

func (c GitClient) VCSType() string {
	return "git"
}

func (c GitClient) VCSRemote() string {
	remote, remoteErr := c.repo.Remote("origin")
	if remoteErr == nil && remote != nil && remote.Config() != nil && len(remote.Config().URLs) > 0 {
		return remote.Config().URLs[0]
	}

	return "local"
}

func (c GitClient) VCSHostServer(remote string) string {
	if remote != "local" {
		// git over ssh
		if strings.HasPrefix(remote, "git@") {
			re := regexp.MustCompile(`(?i)^git@([^:]+):`)
			host := re.FindStringSubmatch(remote)[1]
			return host
		}

		u, err := url.Parse(remote)
		if err != nil {
			log.Warn().Err(err).Msg("error parsing URL")
			return ""
		}

		return u.Host
	}

	return ""
}

func (c GitClient) VCSHostType(server string) string {
	if server == "github.com" {
		return "github"
	} else if server == "gitlab.com" || strings.Contains(server, "gitlab.") {
		return "gitlab"
	} else if len(os.Getenv(toEnvName(server)+"_TYPE")) > 0 {
		return os.Getenv(toEnvName(server) + "_TYPE")
	}

	// host type probes
	if isGitLabInstance(server) {
		return "gitlab"
	} else if isGiteaInstance(server) {
		return "gitea"
	}

	return ""
}

func (c GitClient) VCSRefToInternalRef(ref vcsapi.VCSRef) string {
	if ref.Type == "branch" {
		return REF_HEADS + ref.Value
	} else if ref.Type == "tag" {
		return REF_TAGS + strings.TrimPrefix(ref.Value, "tags/")
	}

	return ref.Hash
}

func (c GitClient) VCSHead() (vcsHead vcsapi.VCSRef, err error) {
	// head reference
	ref, err := c.repo.Head()
	if err != nil {
		return vcsapi.VCSRef{}, err
	}

	if strings.HasPrefix(ref.Name().String(), REF_TAGS) {
		tagName := ref.Name().String()[len(REF_TAGS):]
		return vcsapi.VCSRef{Type: "tag", Value: tagName, Hash: ref.Hash().String()}, nil
	} else if strings.HasPrefix(ref.Name().String(), REF_HEADS) {
		branchName := ref.Name().String()[11:]
		return vcsapi.VCSRef{Type: "branch", Value: branchName, Hash: ref.Hash().String()}, nil
	} else if ref.Name().String() == "HEAD" {
		// detached HEAD, check git reflog for the true reference
		gitRefLogFile := filepath.Join(c.dir, ".git", "logs", "HEAD")
		lastLine := readLastLine(gitRefLogFile)
		return ParseGitRefLogLine(lastLine, ref.Hash().String()), nil
	}

	return vcsapi.VCSRef{}, errors.New("can't determinate repo head")
}

func ParseGitRefLogLine(line string, hash string) vcsapi.VCSRef {
	pattern := regexp.MustCompile(`.*checkout: moving from (?P<FROM>.*) to (?P<TO>.*)$`)
	match := pattern.FindStringSubmatch(line)

	if strings.HasPrefix(match[2], REF_REMOTES+"pull") {
		// handle github merge request as virtual branch
		return vcsapi.VCSRef{Type: "branch", Value: match[2][13:], Hash: hash}
	} else if len(match[2]) == 40 {
		return vcsapi.VCSRef{Type: "branch", Value: match[1], Hash: hash}
	} else if strings.HasPrefix(match[2], REF_TAGS) {
		return vcsapi.VCSRef{Type: "tag", Value: match[2][10:], Hash: hash}
	} else if strings.HasPrefix(match[2], TAGS) { // checkout: moving from develop to tags/v1.0.0
		return vcsapi.VCSRef{Type: "tag", Value: match[2][len(TAGS):], Hash: hash}
	} else {
		return vcsapi.VCSRef{Type: "branch", Value: match[2], Hash: hash}
	}
}

func (c GitClient) GetTags() []vcsapi.VCSRef {
	if c.tags == nil {
		var tags []vcsapi.VCSRef

		iter, err := c.repo.Tags()
		if err == nil {
			iter.ForEach(func(r *plumbing.Reference) error {
				if r.Name().IsTag() {
					t := vcsapi.VCSRef{
						Type:  "tag",
						Value: strings.TrimPrefix(r.Name().String(), REF_TAGS),
						Hash:  r.Hash().String(),
					}
					tags = append(tags, t)
				}

				return nil
			})
		}

		c.tags = tags
	}

	return c.tags
}

func (c GitClient) GetTagsByHash(hash string) []vcsapi.VCSRef {
	if c.tagsByHash == nil {
		c.tagsByHash = make(map[string][]vcsapi.VCSRef)

		iter, err := c.repo.Tags()
		if err == nil {
			iter.ForEach(func(r *plumbing.Reference) error {
				if r.Name().IsTag() {
					t := vcsapi.VCSRef{
						Type:  "tag",
						Value: strings.TrimPrefix(r.Name().String(), REF_TAGS),
						Hash:  r.Hash().String(),
					}

					c.tagsByHash[r.Hash().String()] = append(c.tagsByHash[r.Hash().String()], t)
				}

				return nil
			})
		}
	}

	return c.tagsByHash[hash]
}

func (c GitClient) FindCommitByHash(hash string, includeChanges bool) (vcsapi.Commit, error) {
	commit, err := c.repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return vcsapi.Commit{}, fmt.Errorf("failed to get commit object: %w", err)
	}

	return gitCommitToVCSCommit(commit, c.GetTagsByHash(hash), includeChanges), nil
}

func (c GitClient) FindCommitsBetween(from *vcsapi.VCSRef, to *vcsapi.VCSRef, includeChanges bool, limit int) ([]vcsapi.Commit, error) {
	// from reference
	var fromHash plumbing.Hash
	if from == nil {
		head, headErr := c.repo.Head()
		if headErr != nil {
			return nil, headErr
		}
		fromHash = head.Hash()
	} else {
		fromHash, _ = refToHash(c.repo, c.VCSRefToInternalRef(*from))
	}

	// to reference
	var toHash plumbing.Hash
	if to != nil {
		toHash, _ = refToHash(c.repo, c.VCSRefToInternalRef(*to))
	}

	// commit iterator
	cIter, _ := c.repo.Log(&git.LogOptions{From: fromHash})
	var commits []vcsapi.Commit
	for {
		commit, commitErr := cIter.Next()
		if commitErr != nil {
			break
		}

		// check
		if to != nil && toHash.String() == commit.Hash.String() {
			break
		}

		// limit
		if limit != 0 && len(commits) >= limit {
			break
		}

		commits = append(commits, gitCommitToVCSCommit(commit, c.GetTagsByHash(commit.Hash.String()), includeChanges))
	}

	return commits, nil
}

func (c GitClient) Diff(from *vcsapi.VCSRef, to *vcsapi.VCSRef) ([]vcsapi.VCSDiff, error) {
	var result []vcsapi.VCSDiff

	// from reference
	var fromHash plumbing.Hash
	if from == nil {
		head, headErr := c.repo.Head()
		if headErr != nil {
			return nil, headErr
		}
		fromHash = head.Hash()
	} else {
		fromHash, _ = refToHash(c.repo, c.VCSRefToInternalRef(*from))
	}

	// to reference
	var toHash plumbing.Hash
	if to != nil {
		toHash, _ = refToHash(c.repo, c.VCSRefToInternalRef(*to))
	}

	// find commits
	fromCommit, err := c.repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object for %s: %v", fromHash.String(), err)
	}
	toCommit, err := c.repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object for %s: %v", fromHash.String(), err)
	}

	// diff
	patch, err := fromCommit.Patch(toCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to get patch: %v", err)
	}

	for _, filePatch := range patch.FilePatches() { // except for the changed files this is broken, manually diff the files
		// ignore empty patches (binary files, submodule refs updates)
		if filePatch.IsBinary() || len(filePatch.Chunks()) == 0 {
			continue
		}

		fileFrom, fileTo := filePatch.Files()
		fd := vcsapi.VCSDiff{}
		if fileFrom != nil {
			fd.FileFrom = vcsapi.CommitFile{Name: fileFrom.Path(), Hash: fileFrom.Hash().String(), Mode: fileFrom.Mode().String()}
		}
		if fileTo != nil {
			fd.FileTo = vcsapi.CommitFile{Name: fileTo.Path(), Hash: fileTo.Hash().String(), Mode: fileTo.Mode().String()}
		}

		if fileFrom != nil && fileTo != nil {
			// correct diff
			srcTree, err := fromCommit.Tree()
			if err != nil {
				return nil, fmt.Errorf("failed to get tree from commit: %v", err)
			}
			srcFile, err := srcTree.File(fileFrom.Path())
			if err != nil {
				return nil, fmt.Errorf("failed to get file from tree: %v", err)
			}
			srcContent, err := srcFile.Contents()
			if err != nil {
				return nil, fmt.Errorf("failed to get file content: %v", err)
			}

			dstTree, err := toCommit.Tree()
			if err != nil {
				return nil, fmt.Errorf("failed to get tree from commit: %v", err)
			}
			dstFile, err := dstTree.File(fileTo.Path())
			if err != nil {
				return nil, fmt.Errorf("failed to get file from tree: %v", err)
			}
			dstContent, err := dstFile.Contents()
			if err != nil {
				return nil, fmt.Errorf("failed to get file content: %v", err)
			}

			fd.Lines = vcsapi.Diff(srcContent, dstContent)
		}

		result = append(result, fd)
	}

	return result, nil
}

func (c GitClient) FindLatestRelease(stable bool) (vcsapi.VCSRelease, error) {
	var latestVersion, _ = semver.NewVersion("0.0.0")
	var latest vcsapi.VCSRelease

	tags := c.GetTags()
	for _, tag := range tags {
		version, versionErr := semver.NewVersion(tag.Value)
		if versionErr == nil {
			if version.Compare(latestVersion) > 0 {
				if (stable && len(version.Prerelease()) == 0) || !stable {
					latestVersion = version
					latest = vcsapi.VCSRelease{
						Type:    tag.Type,
						Value:   tag.Value,
						Hash:    tag.Hash,
						Version: version.String(),
					}
				}
			}
		}
	}

	return latest, nil
}

func (c GitClient) CreateBranch(branch string) error {
	// create and checkout new branch
	w, err := c.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
		Force:  true,
	}); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

func (c GitClient) IsClean() (bool, error) {
	// create and checkout new branch
	w, err := c.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	// skip, if no files have changed
	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}
	return status.IsClean(), nil
}

func (c GitClient) UncommittedChanges() ([]string, error) {
	// create and checkout new branch
	w, err := c.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// skip, if no files have changed
	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for file := range status {
		files = append(files, file)
	}

	return files, nil
}

func gitFileActionToText(input merkletrie.Action) string {
	if input == merkletrie.Insert {
		return "create"
	} else if input == merkletrie.Modify {
		return "update"
	} else if input == merkletrie.Delete {
		return "delete"
	}

	return ""
}

func refToHash(repo *git.Repository, ref string) (hash plumbing.Hash, err error) {
	if strings.HasPrefix(ref, "refs/") {
		var pRef *plumbing.Reference
		pRef, err = repo.Reference(plumbing.ReferenceName(ref), true)
		if err != nil {
			return hash, err
		}
		hash = pRef.Hash()
	} else {
		var commit *object.Commit
		commit, err = repo.CommitObject(plumbing.NewHash(ref))
		if err != nil {
			return hash, err
		}
		hash = commit.Hash
	}

	return hash, err
}
