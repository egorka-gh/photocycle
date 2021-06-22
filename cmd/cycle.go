package main

import (
	"fmt"
	clog "log"
	"os"
	"path"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/infrastructure/repo"
	"github.com/egorka-gh/photocycle/job"
	log "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/osext"
	service1 "github.com/kardianos/service"
	group "github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

//demon logger
var dLogger service1.Logger

//
type program struct {
	group     *group.Group
	rep       photocycle.Repository
	interrupt chan struct{}
	quit      chan struct{}
}

//start os demon or console using kardianos
func main() {
	if err := readConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			clog.Output(1, "Start using default setings")
		} else {
			clog.Fatal(err)
			return
		}
	}

	svcConfig := &service1.Config{
		Name:        "Cycle",
		DisplayName: "Cycle Service",
		Description: "Helper service for PhotoCycle",
	}
	prg := &program{}
	s, err := service1.New(prg, svcConfig)
	if err != nil {
		clog.Fatal(err)
	}
	if len(os.Args) > 1 {
		err = service1.Control(s, os.Args[1])
		if err != nil {
			clog.Fatal(err)
		}
		return
	}
	dLogger, err = s.Logger(nil)
	if err != nil {
		clog.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		dLogger.Error(err)
	}
}

func (p *program) Start(s service1.Service) error {
	r, rep, err := initRuner()
	if err != nil {
		return err
	}

	g := &group.Group{}
	p.interrupt = make(chan struct{})
	p.quit = make(chan struct{})
	p.group = g
	p.rep = rep

	runerRunning := make(chan struct{})
	g.Add(func() error {
		return r.Run(runerRunning)
	}, func(error) {
		close(runerRunning)
	})

	//initCancelInterrupt actor
	running := make(chan struct{})
	p.group.Add(
		func() error {
			select {
			case <-p.interrupt:
				return fmt.Errorf("get interrupt signal")
			case <-running:
				return nil
			}
		}, func(error) {
			close(running)
		})

	if service1.Interactive() {
		dLogger.Info("Running in terminal.")
		dLogger.Infof("Valid startup parametrs: %q\n", service1.ControlAction)
	} else {
		dLogger.Info("Starting Netprint service...")
	}
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	//close db cnn
	defer func() {
		if p.rep != nil {
			p.rep.Close()
		}
	}()
	dLogger.Info("Cycle started")
	dLogger.Info(p.group.Run())
	close(p.quit)
}

func (p *program) Stop(s service1.Service) error {
	// Stop should not block. Return with a few seconds.
	dLogger.Info("Cycle Stopping!")
	//interrupt service
	close(p.interrupt)
	//waite service stops
	<-p.quit
	dLogger.Info("Cycle stopped")
	return nil
}

func initRuner() (job.Runer, photocycle.Repository, error) {
	//TODO check settings
	//open database
	rep, err := repo.New(viper.GetString("mysql"), false)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка подключения к базе данных %s", err.Error())
	}
	logger := initLoger(viper.GetString("folders.log"))
	jobs := make([]job.Job, 0, 5)
	if !viper.GetBool("fillBox.off") {
		jobs = append(jobs, job.FillBox())
	}
	if !viper.GetBool("efi.off") {
		jobs = append(jobs, job.PrintedEFI())
	}
	r := job.NewRuner(viper.GetInt("run.interval"), rep, logger, jobs...)
	return r, rep, nil
}

func readConfig() error {
	viper.SetDefault("mysql", "root:3411@tcp(127.0.0.1:3306)/fotocycle_202005?parseTime=true") //MySQL connection string
	viper.SetDefault("folders.log", ".\\log")                                                  //Log folder
	viper.SetDefault("run.interval", 3)                                                        //run interval in mimutes

	folder, err := osext.ExecutableFolder()
	if err != nil {
		folder = "."
	}
	viper.AddConfigPath(folder)
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func initLoger(logPath string) log.Logger {
	var logger log.Logger
	if logPath == "" {
		logger = log.NewLogfmtLogger(os.Stderr)
	} else {
		logPath = path.Join(logPath, "cycle.log")
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    5, // megabytes
			MaxBackups: 5,
			MaxAge:     60, //days
		})
	}
	logger = log.With(logger, "ts", log.DefaultTimestamp) // .DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}
