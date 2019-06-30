package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func InitDBC(dbUser, dbPass, dbName, dbAddr string) error {
	// 数据库连接字符串
	conStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&loc=Asia%%2FShanghai",
		dbUser,
		dbPass,
		dbAddr,
		dbName)

	if err := orm.RegisterDataBase("default", "mysql", conStr, 30); err != nil {
		return err
	}

	// register model.
	orm.RegisterModel(new(Torrent), new(File))

	// sync db.
	return orm.RunSyncdb("default", false, false)
}