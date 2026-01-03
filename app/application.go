package app

import (
	"log"
	"os"
	"path/filepath"

	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var app *Application

type Application struct {
	RootPath       string
	ConfigPath     string
	Config         *config.Config
	ServiceContext *svc.ServiceContext
	MySQL          sqlx.SqlConn
	Gorm           *gorm.DB
}

func GetApplication() *Application {
	return app
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) initRootPath() {
	root := os.Getenv("ROOT_PATH")
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}

		root = wd
	}

	a.RootPath = root
}

func (a *Application) initConfigPath() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/main-api.yaml"
	}

	a.ConfigPath = filepath.Join(a.RootPath, path)
}

func (a *Application) initConfig() {
	a.Config = &config.Config{}
	conf.MustLoad(app.ConfigPath, a.Config)
}

func (a *Application) initServiceContext() {
	a.ServiceContext = svc.NewServiceContext(*a.Config)
}

func (a *Application) Init() {
	a.initRootPath()
	a.initConfigPath()
	a.initConfig()
	a.initServiceContext()
}
