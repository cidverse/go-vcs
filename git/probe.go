package gitclient

import (
	"fmt"
	"net/http"
	"time"
)

// isGitLabInstance probes the server by making a request to the GitLab API.
func isGitLabInstance(host string) bool {
	url := fmt.Sprintf("https://%s/api/v4/projects", host)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// isGiteaInstance checks if the server is running Gitea by querying its API.
func isGiteaInstance(host string) bool {
	url := fmt.Sprintf("https://%s/api/v1/version", host)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
