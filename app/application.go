package app

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel/logger"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/mysql"
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
	Redis          *redis.Redis
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

func (a *Application) initMySQL() {
	a.MySQL = sqlx.NewMysql(a.Config.MySqlConf.Dsn)
	g, err := gorm.Open(mysql.Open(a.Config.MySqlConf.Dsn), &gorm.Config{
		Logger: logger.NewGormLogger(),
	})
	if err != nil {
		log.Fatalf("Failed to create gorm: %v", err)
	}

	db, err := g.DB()
	if err != nil {
		log.Fatalf("Failed to get gorm instance: %v", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	a.Gorm = g
}

func (a *Application) initRedis() {
	a.Redis = redis.MustNewRedis(a.Config.RedisConf)
}

func (a *Application) Init() {
	a.initRootPath()
	a.initConfigPath()
	a.initConfig()
	a.initServiceContext()
	a.initMySQL()
	a.initRedis()
}
