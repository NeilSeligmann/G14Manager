package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/NeilSeligmann/G15Manager/controller"
	"github.com/NeilSeligmann/G15Manager/web"
	suture "github.com/thejerf/suture/v4"

	// "github.com/NeilSeligmann/G15Manager/rpc/server"

	"github.com/NeilSeligmann/G15Manager/supervisor/background"
	"github.com/NeilSeligmann/G15Manager/util"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Compile time injected variables
var (
	Version     = "v0.0.0-dev"
	IsDebug     = "yes"
	logLocation = `G15Manager.log`
	IsPrelogin  = false
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	logLocation = dir + "/" + logLocation

	logger := &lumberjack.Logger{
		Filename:   logLocation,
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	if IsDebug == "no" {
		log.SetOutput(logger)
	} else {
		mw := io.MultiWriter(os.Stdout, logger)
		log.SetOutput(mw)
	}

	log.Printf("G15Manager version: %s\n", Version)

	// Process arguments
	for _, a := range os.Args[1:] {
		if a == "--prelogin" {
			log.Printf("Pre-login mode!")
			IsPrelogin = true
		}
	}

	// Notifier
	notifier := background.NewNotifier()

	// versionChecker, err := background.NewVersionCheck(Version, "zllovesuki/G15Manager", notifier.C)
	// if err != nil {
	// 	log.Fatalf("[supervisor] cannot get version checker")
	// }

	controllerConfig := controller.RunConfig{
		DryRun:     os.Getenv("DRY_RUN") != "",
		NotifierCh: notifier.C,
	}

	dep, err := controller.GetDependencies(controllerConfig)
	if err != nil {
		log.Print(err)
		log.Fatalf("[supervisor] cannot get dependencies\n")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// ------------
	// Controller
	// ------------
	go func() {
		log.Printf("Starting controller...")
		control, _, err := controller.New(controllerConfig, dep)

		// controllerStartErr := <-controllerStartErrCh
		// if controllerStartErr != nil {
		// 	log.Printf("[supervisor] failed to start controller\n")
		// 	log.Fatal(err)
		// 	return
		// }

		if err != nil {
			log.Printf("[supervisor] cannot start controller\n")
			log.Fatal(err)
			return
		}

		controllerSupervisor := suture.New("controllerSupervisor", suture.Spec{})
		controllerSupervisor.Add(control)

		control.Serve(ctx)
	}()

	// Exit if pre-login mode
	if IsPrelogin {
		log.Printf("Pre-login mode finished! Closing...")
		os.Exit(0)
	}

	// ------------
	// Web Server
	// ------------
	web.NewHttpServer(dep)
	// if err != nil {
	// 	log.Fatalf("[supervisor] failed to create HTTP web server: %+v\n", err)
	// }

	// ------------
	// Supervisors
	// ------------

	backgroundSupervisor := suture.New("backgroundSupervisor", suture.Spec{})
	// backgroundSupervisor.Add(versionChecker)
	backgroundSupervisor.Add(notifier)

	rootSupervisor := suture.New("Supervisor", suture.Spec{
		// EventHook: evtHook.Event,
	})
	// rootSupervisor.Add(grpcSupervisor)
	rootSupervisor.Add(backgroundSupervisor)
	// rootSupervisor.Add(NewWeb(grpcServer.GetWebHandler()))

	// -------------
	// Close Signal
	// -------------

	sigc := make(chan os.Signal, 1)

	go func() {
		notifier.C <- util.Notification{
			Message:   "Starting up G15Manager Supervisor",
			Immediate: true,
			Delay:     time.Second * 2,
		}
		supervisorErr := rootSupervisor.Serve(ctx)
		if supervisorErr != nil {
			log.Printf("[supervisor] rootSupervisor returns error: %+v\n", supervisorErr)
			sigc <- syscall.SIGTERM
		}
	}()

	signal.Notify(
		sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	sig := <-sigc
	log.Printf("[supervisor] signal received: %+v\n", sig)

	cancel()
	dep.ConfigRegistry.Close()
	time.Sleep(time.Second) // 1 second for grace period
}
