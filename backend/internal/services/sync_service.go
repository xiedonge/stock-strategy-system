package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SyncOptions carries parameters for AkShare sync.
type SyncOptions struct {
	Symbols  []string
	Mode     string
	StartDate string
	EndDate   string
	MinStart  string
	MinEnd    string
	Period    string
	Limit     int
}

// SyncService runs external sync jobs.
type SyncService struct {
	scriptPath string
	dbPath     string
	timeout    time.Duration
}

// NewSyncService creates a SyncService.
func NewSyncService(scriptPath, dbPath string, timeout time.Duration) *SyncService {
	return &SyncService{scriptPath: scriptPath, dbPath: dbPath, timeout: timeout}
}

// RunAkShareSync executes the AkShare sync script and returns its JSON summary.
func (s *SyncService) RunAkShareSync(options SyncOptions) (map[string]any, error) {
	args := []string{s.scriptPath}
	if len(options.Symbols) > 0 {
		args = append(args, "--symbols", strings.Join(options.Symbols, ","))
	}
	if options.Mode != "" {
		args = append(args, "--mode", options.Mode)
	}
	if options.StartDate != "" {
		args = append(args, "--start-date", options.StartDate)
	}
	if options.EndDate != "" {
		args = append(args, "--end-date", options.EndDate)
	}
	if options.MinStart != "" {
		args = append(args, "--min-start", options.MinStart)
	}
	if options.MinEnd != "" {
		args = append(args, "--min-end", options.MinEnd)
	}
	if options.Period != "" {
		args = append(args, "--period", options.Period)
	}
	if options.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", options.Limit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("DB_PATH=%s", s.dbPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("akshare sync failed: %w: %s", err, string(output))
	}

	var summary map[string]any
	if err := json.Unmarshal(output, &summary); err != nil {
		return nil, fmt.Errorf("parse akshare output: %w: %s", err, string(output))
	}
	return summary, nil
}
