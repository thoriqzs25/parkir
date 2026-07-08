package backups

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type BackupStatus string

const (
	StatusIdle    BackupStatus = "idle"
	StatusRunning BackupStatus = "running"
	StatusSuccess BackupStatus = "success"
	StatusFailed  BackupStatus = "failed"
)

type BackupFile struct {
	Filename  string       `json:"filename"`
	SizeBytes int64        `json:"size_bytes"`
	CreatedAt time.Time    `json:"created_at"`
	Status    BackupStatus `json:"status"`
	Error     string       `json:"error,omitempty"`
}

type Scheduler struct {
	mu          sync.RWMutex
	backupDir   string
	databaseURL string
	status      BackupStatus
	lastRun     *BackupFile
	history     []BackupFile
}

func NewScheduler(backupDir, databaseURL string) *Scheduler {
	s := &Scheduler{
		backupDir:   backupDir,
		databaseURL: databaseURL,
		status:      StatusIdle,
	}
	s.refreshHistory()
	return s
}

func (s *Scheduler) Status() BackupStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Scheduler) LastRun() *BackupFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRun
}

func (s *Scheduler) History() []BackupFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BackupFile, len(s.history))
	copy(result, s.history)
	return result
}

func (s *Scheduler) RunBackup(ctx context.Context) error {
	s.mu.Lock()
	if s.status == StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("backup already in progress")
	}
	s.status = StatusRunning
	s.mu.Unlock()

	timestamp := time.Now().UTC().Format("20060102-150405")
	filename := fmt.Sprintf("parkir-%s.sql.gz", timestamp)
	filePath := filepath.Join(s.backupDir, filename)

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--no-owner",
		"--no-acl",
		"--dbname", s.databaseURL,
	)
	gzip := exec.Command("gzip", "-c")

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		s.setFailed(filename, fmt.Sprintf("pipe error: %v", err))
		return err
	}
	gzip.Stdin = pipe

	outFile, err := os.Create(filePath)
	if err != nil {
		s.setFailed(filename, fmt.Sprintf("create file: %v", err))
		return err
	}
	gzip.Stdout = outFile

	if err := cmd.Start(); err != nil {
		outFile.Close()
		s.setFailed(filename, fmt.Sprintf("pg_dump start: %v", err))
		return err
	}

	if err := gzip.Start(); err != nil {
		outFile.Close()
		s.setFailed(filename, fmt.Sprintf("gzip start: %v", err))
		return err
	}

	cmdErr := cmd.Wait()
	gzipErr := gzip.Wait()
	outFile.Close()

	if cmdErr != nil {
		os.Remove(filePath)
		s.setFailed(filename, fmt.Sprintf("pg_dump failed: %v", cmdErr))
		return cmdErr
	}
	if gzipErr != nil {
		os.Remove(filePath)
		s.setFailed(filename, fmt.Sprintf("gzip failed: %v", gzipErr))
		return gzipErr
	}

	info, err := os.Stat(filePath)
	if err != nil {
		s.setFailed(filename, fmt.Sprintf("stat file: %v", err))
		return err
	}

	bf := BackupFile{
		Filename:  filename,
		SizeBytes: info.Size(),
		CreatedAt: info.ModTime().UTC(),
		Status:    StatusSuccess,
	}

	s.mu.Lock()
	s.status = StatusIdle
	s.lastRun = &bf
	s.history = append([]BackupFile{bf}, s.history...)
	if len(s.history) > 100 {
		s.history = s.history[:100]
	}
	s.mu.Unlock()

	s.rotateOldBackups()
	return nil
}

func (s *Scheduler) setFailed(filename, errMsg string) {
	bf := BackupFile{
		Filename:  filename,
		SizeBytes: 0,
		CreatedAt: time.Now().UTC(),
		Status:    StatusFailed,
		Error:     errMsg,
	}
	s.mu.Lock()
	s.status = StatusIdle
	s.lastRun = &bf
	s.history = append([]BackupFile{bf}, s.history...)
	if len(s.history) > 100 {
		s.history = s.history[:100]
	}
	s.mu.Unlock()
}

func (s *Scheduler) rotateOldBackups() {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-90 * 24 * time.Hour)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(s.backupDir, entry.Name()))
		}
	}
}

func (s *Scheduler) refreshHistory() {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return
	}

	var files []BackupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, BackupFile{
			Filename:  entry.Name(),
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime().UTC(),
			Status:    StatusSuccess,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.After(files[j].CreatedAt)
	})

	s.mu.Lock()
	s.history = files
	if len(s.history) > 0 {
		s.lastRun = &s.history[0]
	}
	s.mu.Unlock()
}
