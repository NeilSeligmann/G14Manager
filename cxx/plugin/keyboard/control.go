//go:generate stringer -type=Level -linecomment
package keyboard

/*
We are looking for a device that looks something like this:
\\?\hid#vid_0b05&pid_1866&mi_02&col01#8&1e16c781&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}
Everything after \hid is just plain old PnP DeviceID, but with "\" replaced with "#"
"vid_X" where X is the vendor ID
"pid_X" where X is the product ID
"mi_X" and "colY" indicates that this is a multi-function, multi-TLC device, and we are looking for a specific column
&1e16c... is the serial number and it *should* be different on each computer
the {uuid} part is generic GUID_DEVINTERFACE_HID: https://docs.microsoft.com/en-us/windows-hardware/drivers/install/guid-devinterface-hid
*/

// #include "virtual.h"
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	// "github.com/NeilSeligmann/G15Manager/rpc/announcement"
	"github.com/NeilSeligmann/G15Manager/system/device"
	"github.com/NeilSeligmann/G15Manager/system/ioctl"
	"github.com/NeilSeligmann/G15Manager/system/keyboard"
	kb "github.com/NeilSeligmann/G15Manager/system/keyboard"
	"github.com/NeilSeligmann/G15Manager/system/persist"
	"github.com/NeilSeligmann/G15Manager/system/plugin"
	"github.com/NeilSeligmann/G15Manager/util"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/micmonay/keybd_event"

	"github.com/karalabe/usb"
)

const (
	persistKey = "KeyboardControl"
)

const (
	brightnessControlByteIndex = 4
)

const (
	brightnessControlBufferLength     = 64
	touchPadToggleControlBufferLength = 64
	initBufferLength                  = 64
)

const (
	kbControlDevice = "mi_00&col01"
)

// TODO: reverse engineer this as well
var (
	brightnessControlBuffer = []byte{
		0x5a, 0xba, 0xc5, 0xc4,
	}
	touchPadToggleControlBuffer = []byte{
		0x5a, 0xf4, 0x6b,
	}
	// initBufs will initialize the keyboard control interface
	// for backlight control and disabling/enabling touchpad.
	// Captured via API Monitor. A lot of HidD_ functions were called
	// (which becomes DeviceIoControl)
	// also referencing https://github.com/flukejones/rog-core/blob/master/kernel-patch/0001-HID-asus-add-support-for-ASUS-N-Key-keyboard-v5.8.patch
	initBufs = [][]byte{
		{
			0x5a, 0x89,
		},
		{
			0x5a, 0x41, 0x53, 0x55, 0x53, 0x20, 0x54, 0x65,
			0x63, 0x68, 0x2e, 0x49, 0x6e, 0x63, 0x2e,
		},
		{
			0x5a, 0x05, 0x20, 0x31, 0x00, 0x08,
		},
	}
)

// Level defines the different level of keybroad brightness
type Level byte

// Brightness level
const (
	OFF    Level = 0x00 // Off
	LOW    Level = 0x01 // Low
	MEDIUM Level = 0x02 // Medium
	HIGH   Level = 0x03 // High
)

// Control allows you to set the hid related functionalities directly.
// The controller is safe for multiple goroutines.
type Control struct {
	Config

	mu                sync.RWMutex
	deviceCtrl        *device.Control
	currentBrightness Level

	queue   chan plugin.Notification
	errChan chan error
}

// Config defines the behavior of Keyboard Control. If DryRun is set to true,
// no actual IOs will be performed. Remap defines the key remapping behavior or
// Fn+ArrowLeft/ArrowRight (see system/keyboard) to standard key scancode.
type Config struct {
	DryRun          bool
	Remap           map[uint32]uint16
	RogKey          []string `json:"rogKey"`
	BrightnessLevel byte     `json:"brightnessLevel"`
}

var _ plugin.Plugin = &Control{}

// NewControl checks if the computer has the hid control interface, and returns a control interface if it does
func NewControl(config Config) (*Control, error) {
	devices, err := usb.EnumerateHid(kb.VendorID, kb.ProductID)
	if err != nil {
		return nil, err
	}
	var path string
	for _, device := range devices {
		if strings.Contains(device.Path, kbControlDevice) {
			path = device.Path
		}
	}
	if path == "" {
		return nil, fmt.Errorf("kbCtrl: Keyboard control interface not found")
	}

	ctrl, err := device.NewControl(device.Config{
		DryRun:      config.DryRun,
		Path:        path,
		ControlCode: ioctl.HID_SET_FEATURE,
	})
	if err != nil {
		return nil, err
	}

	return &Control{
		Config:            config,
		deviceCtrl:        ctrl,
		currentBrightness: OFF,
		queue:             make(chan plugin.Notification),
		errChan:           make(chan error),
	}, nil
}

// Initialize will send initialization buffer to the keyboard control device.
// Note: This should be called prior to calling any control methods, and after
// ACPI resume.
func (c *Control) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("kbCtrl: initializaing hid interface")
	for _, buf := range initBufs {
		initBuf := make([]byte, initBufferLength)
		copy(initBuf, buf)
		_, err := c.deviceCtrl.Write(initBuf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Control) loop(haltCtx context.Context, cb chan<- plugin.Callback) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("kbCtrl: loop panic %+v\n", err)
			c.errChan <- err.(error)
		}
	}()

	for {
		select {
		case t := <-c.queue:
			switch t.Event {
			case plugin.EvtKeyboardFn:
				keycode, ok := t.Value.(uint32)
				if !ok {
					continue
				}
				switch keycode {
				case keyboard.KeyTpadToggle:
					if err := c.ToggleTouchPad(); err != nil {
						c.errChan <- err
					} else {
						cb <- plugin.Callback{
							Event: plugin.CbNotifyToast,
							Value: util.Notification{
								Message: "Toggle Disable/Enable Touchpad",
								Delay:   time.Second,
							},
						}
					}
				case keyboard.KeyFnDown:
					if err := c.BrightnessDown(); err != nil {
						c.errChan <- err
					} else {
						cb <- plugin.Callback{
							Event: plugin.CbNotifyToast,
							Value: util.Notification{
								Message:   fmt.Sprintf("Keyboard Brightness Down: %s", c.CurrentBrightness()),
								Delay:     time.Millisecond * 500,
								Immediate: true,
							},
						}
						cb <- plugin.Callback{
							Event: plugin.CbPersistConfig,
						}
					}
				case keyboard.KeyFnUp:
					if err := c.BrightnessUp(); err != nil {
						c.errChan <- err
					} else {
						cb <- plugin.Callback{
							Event: plugin.CbNotifyToast,
							Value: util.Notification{
								Message:   fmt.Sprintf("Keyboard Brightness Up: %s", c.CurrentBrightness()),
								Delay:     time.Millisecond * 500,
								Immediate: true,
							},
						}
						cb <- plugin.Callback{
							Event: plugin.CbPersistConfig,
						}
					}
				case keyboard.KeyFnF4:
					kb, err := keybd_event.NewKeyBonding()
					if err != nil {
						panic(err)
					}

					kb.SetKeys(keybd_event.VK_MEDIA_PLAY_PAUSE)
					err = kb.Launching()
					if err != nil {
						panic(err)
					}

					// case keyboard.KeyFnLeft, keyboard.KeyFnRight:
					// c.EmulateKeyPress(4274)
					// if remap, ok := c.Config.Remap[keycode]; ok {
					// 	c.EmulateKeyPress(remap)
					// }
				}
			case plugin.EvtACPIResume:
				log.Println("kbCtrl: reinitialize kbCtrl")
				c.errChan <- c.Initialize()
			case plugin.EvtACPISuspend:
				log.Println("kbCtrl: turning off keyboard backlight")
				c.errChan <- c.SetBrightness(OFF)

			case plugin.EvtSentinelUtilityKey:
				counter, ok := t.Value.(int64)
				if !ok {
					continue
				}
				if int(counter) <= len(c.Config.RogKey) {
					var err error
					cmd := c.Config.RogKey[counter-1]
					log.Printf("[controller] Running: %s\n", cmd)

					if cmd == "$webclient" {
						cmd = "http://localhost:34453"
					}

					if govalidator.IsURL(cmd) {
						err = exec.Command("rundll32", "url.dll,FileProtocolHandler", cmd).Start()
					} else {
						err = run("cmd.exe", "/C", cmd)
					}

					if err != nil {
						log.Println("Failed to run command: \"" + cmd + "\"")
						log.Println(err)
					}
				}
			}
		case <-haltCtx.Done():
			log.Println("kbCtrl: exiting Plugin run loop")
			return
		}
	}
}

// Run satifies system/plugin.Plugin
func (c *Control) Run(haltCtx context.Context, cb chan<- plugin.Callback) <-chan error {
	log.Println("kbCtrl: Starting queue loop")

	go c.loop(haltCtx, cb)

	return c.errChan
}

// Notify satifies system/plugin.Plugin
func (c *Control) Notify(t plugin.Notification) {
	c.queue <- t
}

// CurrentBrightness returns current brightness Level
func (c *Control) CurrentBrightness() Level {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Level(c.Config.BrightnessLevel)
}

// SetBrightness change the keyboard backlight directly
func (c *Control) SetBrightness(v Level) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	inputBuf := make([]byte, brightnessControlBufferLength)
	copy(inputBuf, brightnessControlBuffer)
	inputBuf[brightnessControlByteIndex] = byte(v)

	_, err := c.deviceCtrl.Write(inputBuf)
	if err != nil {
		return err
	}

	c.Config.BrightnessLevel = byte(v)

	return nil
}

// BrightnessUp increases the keyboard backlight by one level
func (c *Control) BrightnessUp() error {
	var targetLevel Level
	switch Level(c.Config.BrightnessLevel) {
	case OFF:
		targetLevel = LOW
	case LOW:
		targetLevel = MEDIUM
	case MEDIUM:
		targetLevel = HIGH
	default:
		return nil
	}
	return c.SetBrightness(targetLevel)
}

// BrightnessDown decreases the keyboard backlight by one level
func (c *Control) BrightnessDown() error {
	var targetLevel Level
	switch Level(c.Config.BrightnessLevel) {
	case HIGH:
		targetLevel = MEDIUM
	case MEDIUM:
		targetLevel = LOW
	case LOW:
		targetLevel = OFF
	default:
		return nil
	}
	return c.SetBrightness(targetLevel)
}

// ToggleTouchPad will toggle enabling/disabling the touchpad
func (c *Control) ToggleTouchPad() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	inputBuf := make([]byte, touchPadToggleControlBufferLength)
	copy(inputBuf, touchPadToggleControlBuffer)

	_, err := c.deviceCtrl.Write(inputBuf)
	if err != nil {
		return err
	}

	// I don't think we have a way of checking if the touchpad is disabled/enabled

	return nil
}

// EmulateKeyPress will emulate a keypress via SendInput() scancode.
// Note: some applications using DirectInput may not register this.
func (c *Control) EmulateKeyPress(keyCode uint16) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if C.SendKeyPress(C.ushort(keyCode)) != 0 {
		return fmt.Errorf("kbCtrl: cannot emulate key press")
	}

	return nil
}

var _ persist.Registry = &Control{}

// Name satisfies persist.Registry
func (c *Control) Name() string {
	return persistKey
}

// Value satisfies persist.Registry
func (c *Control) Value() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, _ := json.MarshalIndent(c.Config, "", "")
	return file
}

// Load satisfies persist.Registry
// TODO: check if the input is actually valid
func (c *Control) Load(v []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(v) == 0 {
		return nil
	}

	// Load saved data
	json.Unmarshal(v, &c.Config)

	return nil
}

// Apply satisfies persist.Registry
func (c *Control) Apply() error {
	// mutex already in setBrightness
	return c.SetBrightness(Level(c.Config.BrightnessLevel))
}

// Close satisfied persist.Registry
func (c *Control) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.deviceCtrl.Close()
}

func run(commands ...string) error {
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	return cmd.Start()
}

func (c *Control) GetWSInfo() gin.H {
	return gin.H{
		"currentBrightness": Level(c.Config.BrightnessLevel),
		"rogKey":            c.Config.RogKey,
	}
}

func (c *Control) HandleWSMessage(ws *websocket.Conn, action int, value string) {
	fmt.Printf("HandleWSMessage - Keyboard")
	switch action {
	// Brightness
	case 0:
		i, _ := strconv.Atoi(value)
		c.SetBrightness(Level(i))
	// Rog Key
	case 1:
		arr := strings.Split(value, ",")
		c.Config.RogKey = arr
	// Toggle Touchpad
	case 2:
		c.ToggleTouchPad()
	}
}
