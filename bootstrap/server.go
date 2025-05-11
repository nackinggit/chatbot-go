package bootstrap

import (
	"flag"
	"fmt"
	"os"

	"com.imilair/chatbot/pkg/config"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	gapp app
	Conf config.Config
)

type app struct {
	gin *gin.Engine
}

var (
	ConfigPath string
)

func init() {
	flag.StringVar(&ConfigPath, "configpath", "./configs", "application config path")
}

func (a *app) Init() {
	slogger, _ := zap.NewProduction()
	defer slogger.Sync()
	sugar := slogger.Sugar()
	if !flag.Parsed() {
		flag.Parse()
	}
	version := Version{}
	version.Init()

	if _, err := os.Stat(ConfigPath); err == nil {
		viper.AddConfigPath(ConfigPath)
	} else {
		sugar.Warnf("application config path not setting, use default : %s", "./configs")
		viper.AddConfigPath("./configs")
	}
	viper.SetConfigName("application")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %v", err))
	} else {
		env := viper.GetString("env")
		if env == "" {
			env = "default"
			sugar.Infof("application env not setting, use default : %s", env)
			viper.Set("env", env)
		} else {
			viper.SetConfigFile(fmt.Sprintf("application-%s", env))
			if err := viper.MergeInConfig(); err != nil {
				sugar.Warnf("application env config not setting : %s", env)
				panic(fmt.Errorf("application init err : %v", err))
			}

			if err := viper.MergeInConfig(); err != nil {
				sugar.Warnf("application env config not setting : %s", env)
				panic(fmt.Errorf("application init err : %v", err))
			}

			if err := viper.Unmarshal(&Conf); err != nil {
				panic(fmt.Errorf("application init err : %v", err))
			}
		}
	}
}
