package models
import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/astaxie/beego"
	"fmt"
)

func init() {

	if err := InitDBC(beego.AppConfig.String("DB.User"),
		beego.AppConfig.String("DB.Password"),
		beego.AppConfig.String("DB.Name"),
		beego.AppConfig.String("DB.Addr"));
	err != nil {
		panic(err.Error())
	}
}

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
	orm.RunSyncdb("default", false, false)
	return nil
}