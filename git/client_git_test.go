package gitclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitRefLogLine_Tag(t *testing.T) {
	vcsRef := ParseGitRefLogLine("0000000000000000000000000000000000000000 1cafbbdb80ce27304ac92a9e2fde6c3df8119a19 runner <runner@fv-az554-304.(none)> 1679700466 +0000\tcheckout: moving from master to refs/tags/v2.0.0-alpha.1", "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19")
	assert.Equal(t, "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19", vcsRef.Hash)
	assert.Equal(t, "tag", vcsRef.Type)
	assert.Equal(t, "v2.0.0-alpha.1", vcsRef.Value)
}

func TestParseGitRefLogLine_LocalBranch(t *testing.T) {
	vcsRef := ParseGitRefLogLine("0000000000000000000000000000000000000000 1cafbbdb80ce27304ac92a9e2fde6c3df8119a19 runner <runner@fv-az554-304.(none)> 1679700466 +0000\tcheckout: moving from master to feature-branch", "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19")
	assert.Equal(t, "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19", vcsRef.Hash)
	assert.Equal(t, "branch", vcsRef.Type)
	assert.Equal(t, "feature-branch", vcsRef.Value)
}

func TestParseGitRefLogLine_Hash(t *testing.T) {
	vcsRef := ParseGitRefLogLine("0000000000000000000000000000000000000000 1cafbbdb80ce27304ac92a9e2fde6c3df8119a19 runner <runner@fv-az554-304.(none)> 1679700466 +0000\tcheckout: moving from master to 1cafbbdb80ce27304ac92a9e2fde6c3df8119a19", "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19")
	assert.Equal(t, "1cafbbdb80ce27304ac92a9e2fde6c3df8119a19", vcsRef.Hash)
	assert.Equal(t, "branch", vcsRef.Type)
	assert.Equal(t, "master", vcsRef.Value)
}
