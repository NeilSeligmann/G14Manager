package controller

import (
	"fmt"

	"github.com/NeilSeligmann/G15Manager/cxx/plugin/aidenoise"
	"github.com/NeilSeligmann/G15Manager/cxx/plugin/gpu"
	"github.com/NeilSeligmann/G15Manager/cxx/plugin/keyboard"
	"github.com/NeilSeligmann/G15Manager/cxx/plugin/rr"
	"github.com/NeilSeligmann/G15Manager/cxx/plugin/volume"

	// "github.com/NeilSeligmann/G15Manager/rpc/announcement"
	"github.com/NeilSeligmann/G15Manager/system/atkacpi"
	"github.com/NeilSeligmann/G15Manager/system/battery"
	"github.com/NeilSeligmann/G15Manager/system/persist"
	"github.com/NeilSeligmann/G15Manager/system/plugin"
	"github.com/NeilSeligmann/G15Manager/system/power"
	"github.com/NeilSeligmann/G15Manager/system/thermal"
	"github.com/NeilSeligmann/G15Manager/util"

	"github.com/pkg/errors"
)

// RunConfig contains the start up configuration for the controller
type RunConfig struct {
	DryRun     bool
	PreLogin   bool
	NotifierCh chan util.Notification
}

type Dependencies struct {
	WMI            atkacpi.WMI
	Keyboard       *keyboard.Control
	Battery        *battery.ChargeLimit
	Volume         *volume.Control
	Thermal        *thermal.Control
	GPU            *gpu.Control
	RR             *rr.Control
	AIDenoise      *aidenoise.Control
	ConfigRegistry persist.ConfigRegistry
	// Updatable      []announcement.Updatable
}

func GetDependencies(conf RunConfig) (*Dependencies, error) {

	wmi, err := atkacpi.NewWMI(conf.DryRun)
	if err != nil {
		return nil, err
	}

	var config persist.ConfigRegistry

	if conf.DryRun {
		config, _ = persist.NewDryRegistryHelper()
	} else {
		config, _ = persist.NewRegistryConfigHelper()
	}

	// TODO: make powercfg dryrun-able as well
	powercfg, err := power.NewCfg()
	if err != nil {
		return nil, err
	}

	thermalCfg := thermal.Config{
		WMI:      wmi,
		PowerCfg: powercfg,
		Profiles: thermal.GetDefaultThermalProfiles(),
	}

	thermal, err := thermal.NewControl(thermalCfg)
	if err != nil {
		return nil, err
	}

	battery, err := battery.NewChargeLimit(wmi)
	if err != nil {
		return nil, err
	}

	kbCtrl, err := keyboard.NewControl(keyboard.Config{
		DryRun: conf.DryRun,
		RogKey: []string{"Taskmgr.exe"},
	})
	if err != nil {
		return nil, err
	}

	volCtrl, err := volume.NewVolumeControl(conf.DryRun)
	if err != nil {
		return nil, err
	}

	gpuCtrl, err := gpu.NewGPUControl(conf.DryRun)
	if err != nil {
		return nil, err
	}

	rrCtrl, err := rr.NewRRControl(conf.DryRun)
	if err != nil {
		return nil, err
	}

	var aiDenoiseCtrl *aidenoise.Control

	if !conf.PreLogin {
		aiDenoiseCtrl, err = aidenoise.NewAIDenoise(conf.DryRun)
		if err != nil {
			return nil, err
		}

		config.Register(aiDenoiseCtrl)
	}

	config.Register(kbCtrl)
	config.Register(battery)
	config.Register(thermal)

	// updatable := []announcement.Updatable{
	// 	thermal,
	// 	kbCtrl,
	// }

	return &Dependencies{
		WMI:            wmi,
		Keyboard:       kbCtrl,
		Battery:        battery,
		Volume:         volCtrl,
		Thermal:        thermal,
		GPU:            gpuCtrl,
		RR:             rrCtrl,
		AIDenoise:      aiDenoiseCtrl,
		ConfigRegistry: config,
		// Updatable:      updatable,
	}, nil
}

// New returns a Controller to be ran
func New(conf RunConfig, dep *Dependencies) (*Controller, chan error, error) {

	if dep == nil {
		return nil, nil, fmt.Errorf("nil Dependencies is invalid")
	}
	if dep.WMI == nil {
		return nil, nil, errors.New("nil WMI is invalid")
	}
	if dep.ConfigRegistry == nil {
		return nil, nil, errors.New("nil Registry is invalid")
	}
	if conf.NotifierCh == nil {
		return nil, nil, errors.New("nil NotifierCh is invalid")
	}

	startErrorCh := make(chan error, 1)
	control := &Controller{
		Config: Config{
			WMI: dep.WMI,

			Plugins: []plugin.Plugin{
				dep.Keyboard,
				dep.Volume,
				dep.Thermal,
				dep.GPU,
				dep.RR,
				dep.AIDenoise,
			},
			Registry: dep.ConfigRegistry,

			Notifier: conf.NotifierCh,
		},

		workQueueCh:  make(map[uint32]workQueue, 1),
		errorCh:      make(chan error),
		startErrorCh: startErrorCh,

		keyCodeCh:  make(chan uint32, 1),
		acpiCh:     make(chan uint32, 1),
		powerEvCh:  make(chan uint32, 1),
		pluginCbCh: make(chan plugin.Callback, 1),
	}

	return control, startErrorCh, nil
}
