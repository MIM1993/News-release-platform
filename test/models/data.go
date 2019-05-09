package models

import (
	"github.com/astaxie/beego/orm"
	"log"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type People struct {
	ID   int
	Name string
	Pwd  string
	Articles  []*Article `orm:"reverse(many)"`
}

type Article struct {
	ID      int    `orm:"pk;auto"`
	Title   string  `orm:"unique;size(40)"`
	Content string  `orm:"size(500)"`
	Img     string  `orm:"null"`
	Time    time.Time `orm:"type(datetime);auto_now_add"`
	ReadConst int     `orm:"default(0)"`
	ArticleType  *ArticleType  `orm:"rel(fk);null;on_delete(set_null)"`
	Peoples  []*People          `orm:"rel(m2m)"`
}

type ArticleType struct {
	Id             int
	Typename       string     `orm:"size(40);unique"`
	Article       []*Article  `orm:"reverse(many)"`
}

func init() {
	err := orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/db?charset=utf8")
	if err != nil {
		log.Fatal("RegisterDataBase err:", err)
	}

	orm.RegisterModel(new(People),new(Article),new(ArticleType))

	orm.RunSyncdb("default", false, true)

}
