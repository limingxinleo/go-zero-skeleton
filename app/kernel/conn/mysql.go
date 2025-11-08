package conn

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	//"gorm.io/driver/mysql"
	//"gorm.io/gorm"
	"github.com/limingxinleo/go-zero-skeleton/app/config"
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

type MySQLResult[T any] struct {
	Result T
	Err    error
}

func NewMySQLResult[T any](result T, err error) MySQLResult[T] {
	return MySQLResult[T]{
		Result: result,
		Err:    err,
	}
}

func (m MySQLResult[T]) Handle() (*T, error) {
	switch {
	case m.Err == nil:
		return &m.Result, nil
	case errors.Is(m.Err, sqlx.ErrNotFound):
		return nil, nil
	default:
		return nil, m.Err
	}
}
