package daemon

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"syscleaner/pkg/cleaner"
	"syscleaner/pkg/gaming"

	"github.com/kardianos/service"
)

// Status holds daemon status information.
type Status struct {
	Running            bool
	PID                int
	Uptime             time.Duration
	AutoGamingEnabled  bool
	CleanupSchedule    string
	CPUThreshold       float64
	RAMThreshold       float64
	RecentActions      []Action
}

// Action represents a logged daemon action.
type Action struct {
	Time        string
	Description string
}

var (
	startTime     time.Time
	recentActions []Action
	actionsMu     sync.Mutex
	isRunning     bool
)

type program struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	if gaming.IsEnabled() {
		gaming.Disable()
	}
	isRunning = false
	return nil
}

func (p *program) run() {
	startTime = time.Now()
	isRunning = true
	addAction("Daemon started")

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	p.wg.Add(3)
	go p.monitorGameProcesses(ctx)
	go p.scheduledCleanup(ctx)
	go p.monitorResources(ctx)

	<-ctx.Done()
	addAction("Daemon stopping")
}

func (p *program) monitorGameProcesses(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status := gaming.GetStatus()

			if len(status.ActiveGames) > 0 && !gaming.IsEnabled() {
				gaming.Enable(gaming.Config{AutoDetectGames: true, CPUBoost: 80, RAMReserveGB: 2})
				gameNames := ""
				for _, g := range status.ActiveGames {
					if gameNames != "" {
						gameNames += ", "
					}
					gameNames += g.Name
				}
				addAction(fmt.Sprintf("Auto-enabled gaming mode: %s detected", gameNames))
			} else if len(status.ActiveGames) == 0 && gaming.IsEnabled() {
				gaming.Disable()
				addAction("Auto-disabled gaming mode: no games running")
			}
		}
	}
}

func (p *program) scheduledCleanup(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == 3 {
				addAction("Running scheduled cleanup")
				opts := cleaner.CleanOptions{
					TempFiles: true,
					Logs:      true,
					Prefetch:  true,
				}
				result := cleaner.PerformClean(opts)
				addAction(fmt.Sprintf("Cleanup complete: %d files, %s freed",
					result.FilesDeleted, cleaner.FormatBytes(result.SpaceFreed)))
			}
		}
	}
}

func (p *program) monitorResources(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status := gaming.GetStatus()
			if status.CPUUsage > 90 {
				addAction(fmt.Sprintf("Warning: High CPU usage: %.1f%%", status.CPUUsage))
			}
			if status.RAMUsagePercent > 90 {
				addAction(fmt.Sprintf("Warning: High RAM usage: %.1f%%", status.RAMUsagePercent))
			}
		}
	}
}

func addAction(description string) {
	actionsMu.Lock()
	defer actionsMu.Unlock()

	action := Action{
		Time:        time.Now().Format("15:04:05"),
		Description: description,
	}
	recentActions = append(recentActions, action)
	if len(recentActions) > 10 {
		recentActions = recentActions[len(recentActions)-10:]
	}
}

// GetStatus returns current daemon status.
func GetStatus() Status {
	actionsMu.Lock()
	actions := make([]Action, len(recentActions))
	copy(actions, recentActions)
	actionsMu.Unlock()

	var uptime time.Duration
	if isRunning {
		uptime = time.Since(startTime)
	}

	return Status{
		Running:           isRunning,
		PID:               os.Getpid(),
		Uptime:            uptime,
		AutoGamingEnabled: true,
		CleanupSchedule:   "Daily at 3:00 AM",
		CPUThreshold:      90,
		RAMThreshold:      90,
		RecentActions:     actions,
	}
}

func newServiceConfig() *service.Config {
	return &service.Config{
		Name:        "SysCleaner",
		DisplayName: "SysCleaner Optimization Service",
		Description: "Continuous system optimization and auto-gaming mode",
	}
}

// Install installs the Windows service.
func Install() error {
	prg := &program{}
	s, err := service.New(prg, newServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return s.Install()
}

// Uninstall removes the Windows service.
func Uninstall() error {
	prg := &program{}
	s, err := service.New(prg, newServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return s.Uninstall()
}

// Start starts the Windows service.
func Start() error {
	prg := &program{}
	s, err := service.New(prg, newServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return s.Start()
}

// Stop stops the Windows service.
func Stop() error {
	prg := &program{}
	s, err := service.New(prg, newServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return s.Stop()
}

// Restart restarts the Windows service.
func Restart() error {
	if err := Stop(); err != nil {
		// Ignore stop errors - service may not be running
	}
	time.Sleep(2 * time.Second)
	return Start()
}

// RunForeground runs the daemon in foreground mode for testing.
func RunForeground() error {
	prg := &program{}
	s, err := service.New(prg, newServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return s.Run()
}
