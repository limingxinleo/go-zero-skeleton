package conn

import (
	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MySQL sqlx.SqlConn

var Gorm *gorm.DB

func InitMySQL() {
	MySQL = sqlx.NewMysql(config.Conf.MySqlConf.Dsn)
	Gorm, _ = gorm.Open(mysql.Open(config.Conf.MySqlConf.Dsn), &gorm.Config{})
}

func NewMySQL(dsn string) sqlx.SqlConn {
	return sqlx.NewMysql(dsn)
}

func NewGorm(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(config.Conf.MySqlConf.Dsn), &gorm.Config{})
}

func GetMySQL() sqlx.SqlConn {
	return MySQL
}

func GetGorm() *gorm.DB {
	return Gorm
}

func NewGormDryRunSession() *gorm.DB {
	return Gorm.Session(&gorm.Session{DryRun: true})
}
