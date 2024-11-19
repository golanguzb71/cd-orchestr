package models

// GitCloneRequest contains the details for cloning a repository
type GitCloneRequest struct {
	RepoURL         string `json:"repo_url"`
	DestinationPath string `json:"destination_path"`
}

// GitPullRequest contains the details for pulling from a repository
type GitPullRequest struct {
	RepoPath string `json:"repo_path"`
	Branch   string `json:"branch"`
}

// GitPushRequest contains the details for pushing to a repository
type GitPushRequest struct {
	RepoPath      string `json:"repo_path"`
	Branch        string `json:"branch"`
	CommitMessage string `json:"commit_message"`
}

// GitAddRequest contains the details for adding files to the staging area
type GitAddRequest struct {
	RepoPath string `json:"repo_path"`
}

// GitCommitRequest contains the details for committing changes to a repository
type GitCommitRequest struct {
	RepoPath      string `json:"repo_path"`
	CommitMessage string `json:"commit_message"`
}

// GitSwitchBranchRequest contains the details for switching branches
type GitSwitchBranchRequest struct {
	RepoPath string `json:"repo_path"`
	Branch   string `json:"branch"`
}
