package conn

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	//"gorm.io/driver/mysql"
	//"gorm.io/gorm"
	"main/app/config"
)

var MySQL sqlx.SqlConn

//var Gorm *gorm.DB

func InitMySQL() {
	MySQL = sqlx.NewMysql(config.Conf.MySqlConf.Dsn)
	//Gorm, _ = gorm.Open(mysql.Open(config.Conf.MySqlConf.Dsn), &gorm.Config{})
}

//func NewGormDryRunSession() *gorm.DB {
//	return Gorm.Session(&gorm.Session{DryRun: true})
//}
