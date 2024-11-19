package service

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"session-bridge/internal/db"
	models "session-bridge/internal/model"
	"time"
)

type GitService struct {
	redis *db.Redis
}

func NewGitService(redis *db.Redis) *GitService {
	return &GitService{redis: redis}
}

// CloneRepo clones a repository to a specified path on the server
func (s *GitService) CloneRepo(conn *models.ServerConnection, req *models.GitCloneRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("git clone %s %s", req.RepoURL, req.DestinationPath)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

// PullRepo pulls changes from a specified branch
func (s *GitService) PullRepo(conn *models.ServerConnection, req *models.GitPullRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && git pull origin %s", req.RepoPath, req.Branch)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

// PushRepo commits and pushes changes to the repository
func (s *GitService) PushRepo(conn *models.ServerConnection, req *models.GitPushRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && git add . && git commit -m \"%s\" && git push origin %s", req.RepoPath, req.CommitMessage, req.Branch)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

// AddFiles stages files for commit
func (s *GitService) AddFiles(conn *models.ServerConnection, req *models.GitAddRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && git add .", req.RepoPath)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

// CommitChanges commits staged changes with a message
func (s *GitService) CommitChanges(conn *models.ServerConnection, req *models.GitCommitRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && git commit -m \"%s\"", req.RepoPath, req.CommitMessage)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

// SwitchBranch switches to a specified branch
func (s *GitService) SwitchBranch(conn *models.ServerConnection, req *models.GitSwitchBranchRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && git checkout %s", req.RepoPath, req.Branch)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

func (s *GitService) getSSHClient(conn *models.ServerConnection) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: conn.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", conn.IPAddress), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return client, nil
}
