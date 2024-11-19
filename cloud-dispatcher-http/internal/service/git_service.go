package service

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"session-bridge/internal/db"
	models "session-bridge/internal/model"
	"strings"
	"time"
)

type GitService struct {
	redis *db.Redis
}

func NewGitService(redis *db.Redis) *GitService {
	return &GitService{redis: redis}
}

func (s *GitService) CloneRepo(conn *models.ServerConnection, req *models.GitCloneRequest) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}
	defer client.Close()
	urlParts := strings.Split(req.RepoURL, "/")
	repoName := urlParts[len(urlParts)-1]
	repoName = strings.TrimSuffix(repoName, ".git")
	req.DestinationPath = filepath.Join("/home/abdulaziz", repoName)
	checkCmd := fmt.Sprintf("if [ -d \"%s\" ]; then echo \"exists\"; else echo \"not exists\"; fi", req.DestinationPath)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	session.Stderr = &stderr
	var stdout bytes.Buffer
	session.Stdout = &stdout

	err = session.Run(checkCmd)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %v, stderr: %s", err, stderr.String())
	}

	if stdout.String() == "exists\n" {
		return fmt.Errorf("destination path '%s' already exists and is not empty", req.DestinationPath)
	}

	cmd := fmt.Sprintf("git clone %s %s", req.RepoURL, req.DestinationPath)

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	session.Stderr = &stderr
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

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

	cmd := ""
	if req.IsFilled {
		cmd = fmt.Sprintf("cd %s && git add . && git commit -m \"%s\" && git push origin %s", req.RepoPath, req.CommitMessage, req.Branch)
	} else {
		cmd = fmt.Sprintf("cd %s && git push origin %s", req.RepoPath, req.Branch)
	}

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

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
