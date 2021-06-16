package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/egorka-gh/photocycle/infrastructure/api"
	"github.com/egorka-gh/photocycle/infrastructure/repo"
	"github.com/egorka-gh/photocycle/netprint"
	log "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	var offsetStr string
	if len(os.Args) > 1 {
		offsetStr = os.Args[1]
	}
	if offsetStr == "" {
		fmt.Println("Укажите период опроса в часах")
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if offset < 1 {
		offset = 1
	}

	if err != nil {
		fmt.Println("Ошибка преобразования в число")
		return
	}
	fmt.Printf("Период опроса %d часов \n", offset)
	if err := readConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("Start using default setings")
		} else {
			fmt.Println(err.Error())
			return
		}
	}
	//TODO check settings
	sourceID := viper.GetInt("source.id")
	if sourceID == 0 {
		fmt.Println("Не задано ID источника")
		return
	}

	//open database
	rep, err := repo.New(viper.GetString("mysql"), false)
	if err != nil {
		fmt.Printf("Ошибка подключения к базе данных %s\n", err.Error())
		return
	}
	defer rep.Close()

	//logger := initLoger(viper.GetString("folders.log"))
	logger := initLoger("")

	client, err := api.NewClient(http.DefaultClient, viper.GetString("source.url"), viper.GetString("source.appKey"))
	if err != nil {
		fmt.Println(err)
		return
	}
	m := netprint.New(sourceID, offset, client, rep, logger)
	m.Sync(context.Background())
}

func readConfig() error {
	viper.SetDefault("mysql", "root:3411@tcp(127.0.0.1:3306)/fotocycle_cycle?parseTime=true") //MySQL connection string
	viper.SetDefault("source.id", 11)                                                         //photocycle source id
	viper.SetDefault("source.url", "https://fabrika-fotoknigi.ru/api/")                       //photocycle source url
	viper.SetDefault("source.appKey", "e5ea49c386479f7c30f60e52e8b9107b")                     //source site appkey
	viper.SetDefault("folders.log", ".\\log")                                                 //Log folder

	path, err := osext.ExecutableFolder()
	if err != nil {
		path = "."
	}
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func initLoger(logPath string) log.Logger {
	var logger log.Logger
	if logPath == "" {
		logger = log.NewLogfmtLogger(os.Stderr)
	} else {
		path := logPath
		if !os.IsPathSeparator(path[len(path)-1]) {
			path = path + string(os.PathSeparator)
		}
		path = path + "order.log"
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    5, // megabytes
			MaxBackups: 5,
			MaxAge:     60, //days
		})
	}
	logger = log.With(logger, "ts", log.DefaultTimestamp) // .DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}
