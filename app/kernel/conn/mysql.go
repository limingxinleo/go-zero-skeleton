package conn

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"main/app/config"
)

var MySQL sqlx.SqlConn

func InitMySQL() {
	MySQL = sqlx.NewMysql(config.Conf.MySqlConf.Dsn)
}
