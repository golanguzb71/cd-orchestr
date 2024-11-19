package models

type GitCloneRequest struct {
	RepoURL         string `json:"repo_url"`
	DestinationPath string `json:"destination_path"`
}

type GitPullRequest struct {
	RepoPath string `json:"repo_path"`
	Branch   string `json:"branch"`
}

type GitPushRequest struct {
	RepoPath      string `json:"repo_path"`
	Branch        string `json:"branch"`
	CommitMessage string `json:"commit_message"`
	IsFilled      bool   `json:"is_filled"`
}

type GitAddRequest struct {
	RepoPath string `json:"repo_path"`
}

type GitCommitRequest struct {
	RepoPath      string `json:"repo_path"`
	CommitMessage string `json:"commit_message"`
}

type GitSwitchBranchRequest struct {
	RepoPath string `json:"repo_path"`
	Branch   string `json:"branch"`
}
