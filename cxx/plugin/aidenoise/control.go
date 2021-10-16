package aidenoise

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	// "syscall"
	"time"

	"github.com/NeilSeligmann/G15Manager/system/plugin"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	persistKey = "AIDenoise"
)

type Control struct {
	dryRun  bool
	queue   chan plugin.Notification
	errChan chan error

	config *AINoiseConfig

	denoiseCmd      *exec.Cmd
	runningCheck    bool
	foundExecutable bool
	shouldRestart   bool
}

type AINoiseConfig struct {
	Enabled     bool   `json:"enabled"`
	DenoisePath string `json:"denoisePath"`
}

func NewAIDenoise(dryRun bool) (*Control, error) {
	return &Control{
		dryRun:          dryRun,
		queue:           make(chan plugin.Notification),
		errChan:         make(chan error),
		foundExecutable: false,
		config:          defaultConfig(),
	}, nil
}

func defaultConfig() *AINoiseConfig {
	return &AINoiseConfig{
		Enabled:     true,
		DenoisePath: "C:\\Program Files\\ASUS\\ARMOURY CRATE Service\\DenoiseAIPlugin\\ArmouryCrate.DenoiseAI.exe",
	}
}

// Initialize satisfies system/plugin.Plugin
func (c *Control) Initialize() error {
	// go c.executeDenoise()
	c.stopRunningInstances()

	return nil
}

// Run satisfies system/plugin.Plugin
func (c *Control) Run(haltCtx context.Context, cb chan<- plugin.Callback) <-chan error {
	log.Println("denoise: Starting queue loop")

	c.runningCheck = true
	go c.loopRunningCeck()

	go c.loop(haltCtx, cb)

	return c.errChan
}

func (c *Control) loop(haltCtx context.Context, cb chan<- plugin.Callback) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("denoise: loop panic %+v\n", err)
			c.errChan <- err.(error)
		}
	}()

	for {
		select {
		// case evt := <-c.queue:
		case <-c.queue:
			// log.Println("aiDenoise: queue!")
			if c.dryRun {
				log.Println("aiDenoise: dry run, queue ignored")
				continue
			}

		case <-haltCtx.Done():
			c.runningCheck = false
			log.Println("denoise: exiting Plugin run loop")
			return
		}
	}
}

func (c *Control) loopRunningCeck() {
	for {
		if !c.runningCheck {
			c.stopDenoise()
			break
		}

		if c.config.Enabled {
			if !c.isRunning() {
				log.Println("aiDenoise: denoise process not found! Executing denoise...")
				c.executeDenoise()
			} else if c.shouldRestart {
				log.Println("aiDenoise: denoise process waiting to be restarted. Restarting...")
				c.executeDenoise()
			}
		}

		time.Sleep(time.Second * 5)
	}
}

// Notify satisfies system/plugin.Plugin
func (c *Control) Notify(t plugin.Notification) {
	if c.dryRun {
		return
	}

	c.queue <- t
}

func (c *Control) isRunning() bool {
	// return c.denoiseCmd != nil

	if c.denoiseCmd == nil {
		return false
	}

	if c.denoiseCmd.Process == nil {
		return false
	}

	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(c.denoiseCmd.Process.Pid))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmdOutput, _ := cmd.Output()

	output := string(cmdOutput[:])
	splitOutp := strings.Split(output, " ")
	if !(strings.ToLower(splitOutp[1]) == "no") {
		return true
	} else {
		return false
	}
}

func (c *Control) executeDenoise() error {
	c.shouldRestart = false
	log.Printf("denoise: Stopping running denoise instance... (if any)")

	// Stop if already running
	c.stopDenoise()

	// Early return if disabled
	if !c.config.Enabled {
		log.Printf("denoise: Denoise is disabled, will not be executed.")
		return nil
	}

	denoisePath := c.config.DenoisePath

	log.Printf("denoise: Starting Denoise AI...")
	log.Printf("denoise: Executable path is -> \"%s\"", denoisePath)

	// Check if the executable exists
	_, err := os.Stat(denoisePath)
	if os.IsNotExist(err) {
		log.Printf("DenoiseAI executable was not found at \"" + denoisePath + "\"")
		return err
	}

	cmd := exec.Command(denoisePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Run denoise executable
	err = cmd.Start()
	if err != nil {
		log.Printf("Failed to execute Denoise AI:")
		log.Print(err)
		return err
	}

	c.denoiseCmd = cmd

	log.Printf("denoise: Successfully executed denoise!")

	return nil
}

func (c *Control) stopDenoise() error {
	// Stop if already running
	if c.denoiseCmd == nil {
		return nil
	}

	if c.denoiseCmd.Process == nil {
		return nil
	}

	err := c.denoiseCmd.Process.Kill()
	if err != nil {
		return err
	}

	return nil
}

func (c *Control) stopRunningInstances() {
	log.Printf("denoise: Stopping already running instances...")

	denoisePathSplit := strings.Split(c.config.DenoisePath, "\\")
	execName := denoisePathSplit[len(denoisePathSplit)-1]

	cmd := exec.Command("tasklist", "/FO", "CSV", "/FI", "IMAGENAME eq "+execName, "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmdOutput, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to fetch list of running AIDenoise processes.")
		return
	}

	output := string(cmdOutput[:])
	lines := strings.Split(output, "\n")

	if len(lines) < 1 {
		log.Printf("AIDenoise is not running, no need to close old instances.")
		return
	}

	for _, v := range lines {
		if v == "" {
			break
		}

		replaced := strings.ReplaceAll(v, "\"", "")
		split := strings.Split(replaced, ",")
		if len(split) < 2 {
			break
		}

		pid := split[1]

		execCmd := exec.Command("Taskkill", "/F", "/PID", pid)
		execErr := execCmd.Start()

		if execErr == nil {
			log.Printf("Killed existing AIDenoise process with PID: %s", pid)
		} else {
			log.Printf("Failed to kill process with PID: %s", pid)
		}
	}
}

// ---------
// Registry
// ---------

// Name satisfies persist.Registry
func (c *Control) Name() string {
	return persistKey
}

// Value satisfies persist.Registry
func (c *Control) Value() []byte {
	file, _ := json.MarshalIndent(c.config, "", "")
	return file
}

// Load satisfies persist.Registry
func (c *Control) Load(v []byte) error {
	if len(v) == 0 {
		return nil
	}

	loadedConfig := AINoiseConfig{}

	// Load saved data
	json.Unmarshal(v, &loadedConfig)

	if loadedConfig.DenoisePath != c.config.DenoisePath {
		c.shouldRestart = true
	}

	c.config = &loadedConfig

	// Restart if needed
	if c.shouldRestart {
		c.executeDenoise()
	}

	return nil
}

// Apply satisfies persist.Registry
func (c *Control) Apply() error {
	return nil
}

// Close satisfied persist.Registry
func (c *Control) Close() error {
	return nil
}

// -----------
// Web Socket
// -----------
func (c *Control) GetWSInfo() gin.H {
	return gin.H{
		"isRunning":      c.isRunning(),
		"executablePath": c.config.DenoisePath,
		"isEnabled":      c.config.Enabled,
		"pid":            c.denoiseCmd.Process.Pid,
	}
}

func (c *Control) HandleWSMessage(ws *websocket.Conn, action int, value string) {
	switch action {
	// Set enabled/disabled
	case 0:
		// Toggle if empty
		if value == "" {
			c.config.Enabled = !c.config.Enabled
		} else {
			c.config.Enabled = value == "1"
		}
	// Set Executable Path
	case 1:
		log.Printf("denoise: new denoise path: %s", value)
		c.config.DenoisePath = value
	// Reset path to default
	case 2:
		c.config.DenoisePath = defaultConfig().DenoisePath
	}

	// Restart denoise with new config
	c.executeDenoise()
}
