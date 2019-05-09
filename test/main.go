package main

import (
	_ "test/routers"
	"github.com/astaxie/beego"
	_"test/models"
)

func main() {
	beego.AddFuncMap("PerPage",PerPage)
	beego.AddFuncMap("NextPage",NextPage)
	beego.AddFuncMap("AddOne",AddOne)
	beego.Run()
}

func PerPage(pageNum int)int{
	if pageNum == 1{
		return 1
	}
	return pageNum -1
}

func NextPage(pageNum int,pageCount float64)int{
	if pageNum == int(pageCount) {
		return int(pageCount)
	}
	return pageNum+1
}

func AddOne(index int)int{
	return index+1
}