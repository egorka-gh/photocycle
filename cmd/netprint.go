package main

import (
	"fmt"
	clog "log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/api"
	"github.com/egorka-gh/photocycle/netprint"
	"github.com/egorka-gh/photocycle/repo"
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

	sourceID := viper.GetInt("source.id")
	svcConfig := &service1.Config{
		Name:        fmt.Sprintf("Netprint_%d", sourceID),
		DisplayName: fmt.Sprintf("Netprint Service id: %d", sourceID),
		Description: "Netprint service for PhotoCycle",
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
	m, rep, err := initNetprint()
	if err != nil {
		return err
	}

	g := &group.Group{}
	p.interrupt = make(chan struct{})
	p.quit = make(chan struct{})
	p.group = g
	p.rep = rep

	//manager actor
	managerRunning := make(chan struct{})
	g.Add(func() error {
		m.Run(viper.GetInt("sync.interval"), managerRunning)
		return nil
	}, func(error) {
		close(managerRunning)
	})

	//initCancelInterrupt actor
	running := make(chan struct{})
	p.group.Add(
		func() error {
			select {
			case <-p.interrupt:
				return fmt.Errorf("Get interrupt signal")
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
	dLogger.Info("Netprint started")
	dLogger.Info(p.group.Run())
	close(p.quit)
}

func (p *program) Stop(s service1.Service) error {
	// Stop should not block. Return with a few seconds.
	dLogger.Info("Netprint Stopping!")
	//interrupt service
	close(p.interrupt)
	//waite service stops
	<-p.quit
	dLogger.Info("Netprint stopped")
	return nil
}

func initNetprint() (*netprint.Manager, photocycle.Repository, error) {
	//TODO check settings
	sourceID := viper.GetInt("source.id")
	if sourceID == 0 {
		return nil, nil, fmt.Errorf("Не задано ID источника")
	}
	//open database
	rep, err := repo.New(viper.GetString("mysql"), false)
	if err != nil {
		return nil, nil, fmt.Errorf("Ошибка подключения к базе данных %s", err.Error())
	}
	logger := initLoger(viper.GetString("folders.log"))
	// use custom http client
	c := &http.Client{
		Timeout: time.Second * 40,
	}
	client, err := api.NewClient(c, viper.GetString("source.url"), viper.GetString("source.appKey"))
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	m := netprint.New(sourceID, viper.GetInt("sync.offset"), client, rep, logger)
	return m, rep, nil
}

func readConfig() error {
	viper.SetDefault("mysql", "root:3411@tcp(127.0.0.1:3306)/fotocycle_cycle?parseTime=true") //MySQL connection string
	viper.SetDefault("source.id", 11)                                                         //photocycle source id
	viper.SetDefault("source.url", "https://fabrika-fotoknigi.ru/api/")                       //photocycle source url
	viper.SetDefault("source.appKey", "e5ea49c386479f7c30f60e52e8b9107b")                     //source site appkey
	viper.SetDefault("folders.log", ".\\log")                                                 //Log folder
	viper.SetDefault("sync.interval", 20)                                                     //sunc interval in mimutes
	viper.SetDefault("sync.offset", 3)                                                        //sunc offset in hours

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
		logPath = path.Join(logPath, "netprint.log")
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
