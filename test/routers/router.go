package routers

import (
	"test/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {

//-------------------------陆游过滤器-------------------------------------
                               //匹配方式   过滤位置（一共有5个位置）  回调函数
    beego.InsertFilter("/article/*",beego.BeforeExec,filterFunc)

//---------------------------用户陆游---------------------------------------------
	//注册
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")

	//登陆
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")

    //用户退出
    beego.Router("/logout",&controllers.UserController{},"get:Logout")


//---------------------------文章陆游-------------------------------------------------------
	//进入首页
	beego.Router("/article/index",&controllers.ArticleController{},"get,post:ShowIndex")
    //
	//添加文章
	beego.Router("/article/addArticle",&controllers.ArticleController{},"get:AddArticle;post:HandleAddArticle")

	//首页、末页、上一页、下一页
	beego.Router("/article/ShowArticleList",&controllers.ArticleController{},"get:ShowIndex")

	//查看详情
	beego.Router("/article/showContent",&controllers.ArticleController{},"get:ShowContent")

	//编辑
	beego.Router("/article/edit",&controllers.ArticleController{},"get:ShowEdit;post:UpdateArticle")

	//删除
	beego.Router("/article/delete",&controllers.ArticleController{},"get:HandleDelete")

	//添加文章类型
	beego.Router("/article/addType",&controllers.ArticleController{},"get:ShowArticleType;post:HandleAddType")

	//删除文章类型
	beego.Router("/article/deletetype",&controllers.ArticleController{},"get:DeleteType")
}


func filterFunc(ctx *context.Context){
	userName := ctx.Input.Session("userName")
	if userName == nil{
		ctx.Redirect(302,"/login")
		return
	}
}