package service

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"regexp"
	"session-bridge/internal/db"
	models "session-bridge/internal/model"
	"strconv"
	"strings"
	"time"
)

type FileFolderService struct {
	redis *db.Redis
}

func NewFileFolderService(redis *db.Redis) *FileFolderService {
	return &FileFolderService{redis: redis}
}

func (s *FileFolderService) getSSHClient(conn *models.ServerConnection) (*ssh.Client, error) {
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

func (s *FileFolderService) ListPath(conn *models.ServerConnection, path string) ([]models.FileInfo, error) {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	if path == "" {
		path = "~"
	}

	cmd := fmt.Sprintf("ls -la %s", path)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}

	return parseLsOutput(string(output), path)
}

func (s *FileFolderService) CreateFolder(conn *models.ServerConnection, req *models.CreateFolderRequest, itemType string) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	fullPath := filepath.Join(req.Path, req.Name)
	var cmd string
	fmt.Println(itemType == "file")
	if itemType == "file" {
		cmd = fmt.Sprintf("touch %s", fullPath)
	} else {
		cmd = fmt.Sprintf("mkdir -p %s", fullPath)
	}

	return session.Run(cmd)
}

func (s *FileFolderService) EditPath(conn *models.ServerConnection, req *models.EditRequest, newFileName string) error {
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

	newFilePath := filepath.Join(req.NewPath, newFileName)

	cmd := fmt.Sprintf("mv %s %s", req.OldPath, newFilePath)

	var stderr bytes.Buffer
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, stderr: %s", cmd, err, stderr.String())
	}

	return nil
}

func (s *FileFolderService) DeletePath(conn *models.ServerConnection, path string) error {
	client, err := s.getSSHClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("rm -rf %s", path)
	return session.Run(cmd)
}

func parseLsOutput(output string, basePath string) ([]models.FileInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return []models.FileInfo{}, nil
	}

	var files []models.FileInfo
	pattern := regexp.MustCompile(`^([d-][rwx-]{9})\s+(\d+)\s+(\w+)\s+(\w+)\s+(\d+)\s+(\w+\s+\d+\s+(?:\d{2}:?\d{2}|\d{4}))\s+(.+)$`)

	for _, line := range lines[1:] {
		matches := pattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) != 8 {
			continue
		}

		perms := matches[1]
		isDir := perms[0] == 'd'
		name := matches[7]

		if name == "." || name == ".." {
			continue
		}

		size, err := strconv.ParseInt(matches[5], 10, 64)
		if err != nil {
			size = 0
		}

		modTime, err := parseModTime(matches[6])
		if err != nil {
			modTime = time.Now().Format(time.RFC3339)
		}

		fullPath := filepath.Join(basePath, name)
		if basePath == "~" {
			fullPath = "~/" + name
		}

		fileInfo := models.FileInfo{
			Name:       name,
			Path:       fullPath,
			Size:       size,
			IsDir:      isDir,
			ModTime:    modTime,
			Permission: parsePermissions(perms),
			Owner:      matches[3],
			Group:      matches[4],
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

func parseModTime(modTimeStr string) (string, error) {
	layouts := []string{
		"Jan 2 15:04",
		"Jan 2 2006",
		"Jan 02 15:04",
		"Jan 02 2006",
	}

	currentYear := time.Now().Year()
	var parsedTime time.Time
	var err error

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, modTimeStr)
		if err == nil {
			if !strings.Contains(layout, "2006") {
				parsedTime = parsedTime.AddDate(currentYear, 0, 0)

				if parsedTime.After(time.Now()) {
					parsedTime = parsedTime.AddDate(-1, 0, 0)
				}
			}
			return parsedTime.Format(time.RFC3339), nil
		}
	}

	return "", err
}

func parsePermissions(perms string) string {
	if len(perms) != 10 {
		return "unknown"
	}

	var result strings.Builder

	switch perms[0] {
	case 'd':
		result.WriteString("directory")
	case '-':
		result.WriteString("file")
	case 'l':
		result.WriteString("link")
	default:
		result.WriteString("special")
	}

	result.WriteString(" [")

	if perms[1] == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if perms[2] == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if perms[3] == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	result.WriteString("|")

	if perms[4] == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if perms[5] == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if perms[6] == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	result.WriteString("|")

	if perms[7] == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if perms[8] == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if perms[9] == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	result.WriteString("]")

	return result.String()
}
