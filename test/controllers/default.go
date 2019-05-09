package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"test/models"
	"fmt"
	"path"
	"time"
	"math"
	"strconv"
	"encoding/base64"
	"github.com/gomodule/redigo/redis"
	"bytes"
	"encoding/gob"
)

//-------------------控制器------------------------------------------------------
//用户控制器
type UserController struct {
	beego.Controller
}
//文章控制器
type ArticleController struct {
	beego.Controller
}


//------------------用户操作函数--------------------------------------------------

//用户注册页面展示
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//用户注册操作
func (this *UserController) HandleRegister() {
	userName := this.GetString("userName")
	pwd := this.GetString("password")

	if userName == "" || pwd == "" {
		fmt.Println("数据不完整")
		this.TplName = "register.html"
		return
	}

	o := orm.NewOrm()

	var user models.People

	user.Name = userName
	user.Pwd = pwd

	Id, err := o.Insert(&user)
	if err != nil {
		fmt.Println("注册失败")
		this.TplName = "register.html"
		return
	}
	fmt.Println("注册Id为：", Id)

	//渲染
	//this.TplName="login.html"
	//跳转
	this.Redirect("/login", 302)
}

//用户登录展示
func (this *UserController) ShowLogin() {
	//获取cookie
	userName := this.Ctx.GetCookie("userName")
	//解密密文
	dec,_ := base64.StdEncoding.DecodeString(userName)
	//记住用户名，向登陆界面传用户名
	if userName!=""{
		this.Data["userName"]=string(dec)
		this.Data["checked"]="checked"
	}else {
		this.Data["userName"]=""
		this.Data["checked"]=""
	}

	this.TplName = "login.html"
}

//用户登陆操作
func (this *UserController) HandleLogin() {
	userName := this.GetString("userName")
	pwd := this.GetString("password")

	if userName == "" || pwd == "" {
		fmt.Println("登陆数据不完整")
		this.TplName = "login.html"
		return
	}

	o := orm.NewOrm()

	var user models.People

	user.Name = userName

	err := o.Read(&user, "Name")
	if err != nil {
		fmt.Println("用户名不正确")
		this.TplName = "login.html"
		return
	}

	if user.Pwd != pwd {
		fmt.Println("用户密码错误")
		this.TplName = "login.html"
		return
	}

	//获取记住用户名点击匡属性
	remember := this.GetString("remember")
	//加密用户名
	enc :=base64.StdEncoding.EncodeToString([]byte(userName))
	//如果点击了记住用户名匡，设置cookie
	if remember == "on"{
		this.Ctx.SetCookie("userName",enc,60)
	}else {
		this.Ctx.SetCookie("userName",userName,-1)
	}

	//设置session存储用户名
	this.SetSession("userName",userName)

	//this.Ctx.WriteString("登陆成功！")
	this.Redirect("/article/index", 302)

}

//用户退出
func (this *UserController)Logout(){
	//删除session
	this.DelSession("userName")

	//跳转到登陆页面
	this.Redirect("/login",302)
}


//---------------------------文章操作函数---------------------------------------

//展示首页页面【重点】
func (this *ArticleController) ShowIndex() {
	//校验登陆状态
	userName := this.GetSession("userName")
	//上面函数返回值为接口类型，空值为nil
	if userName == nil{
		this.Redirect("/login",302)
		return
	}

	//向首页传递用户名
	this.Data["userName"]=userName.(string)//类型断言


//--------------------------------------------------------
	//获取选中的类型
	typeName := this.GetString("select")

	//获取文件数据，展示到页面上
	o := orm.NewOrm()
	//指定要查询的数据库表，用QueryTable函数
	query := o.QueryTable("Article")

	//获取记录总数
	var count int64
	if typeName == ""{
		count,_=query.RelatedSel("ArticleType").Count()
	}else {
		count,_=query.RelatedSel("ArticleType").Filter("ArticleType__Typename",typeName).Count()
	}

	//确定每页记录数量
	pageSize := 2

	//总页数=总记录数/每页记录数（向上取整数）
	pageCount := math.Ceil(float64(count) / float64(pageSize))

	//获取首页和末页数据
	//获取页码
	pageNum, err := this.GetInt("pageNum", 1)
	if err != nil {
		fmt.Println("获取页码错误")
		this.Data["errmsg"] = "获取页码错误"
		this.Layout="Layout.html"
		this.TplName = "index.html"
		return
	}

	//获取起始数据的位置
	start := (pageNum - 1) * pageSize

	//创建一个对象类型切片
	var articles []models.Article

	//获取对应页数据
	//多表查询是惰性查询，必须关联表
	//where ArticleType.typeName = typename   filter相当于where条件语句
	if typeName == "" {
		query.Limit(pageSize, start).RelatedSel("ArticleType").All(&articles)

	} else {
		query.Limit(pageSize, start).RelatedSel("ArticleType").Filter("ArticleType__Typename", typeName).All(&articles)

	}

	//将读取过的数据存放在redis，首次读取从mysql读取
	//连接redis数据库
	Conn,err := redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		fmt.Println("redis连接失败")
		return
	}

	//创建文章类型对像切片
	var articleTypes []models.ArticleType
	//从redis数据库读取数据
	repy,err := Conn.Do("get","newsWeb")
	//助手函数转换数据
	result,_:=redis.Bytes(repy,err)
	//判断是否读取到了数据
	if len(result)==0{
		//获取文章类型数据
		o.QueryTable("ArticleType").All(&articleTypes)

		//在redis数据库创建数据
		//首先序列化
		//创建容器
		var buffer bytes.Buffer
		//创建编码器
		enc := gob.NewEncoder(&buffer)
		//进行编码
		enc.Encode(articleTypes)
		//存入redis数据库
		Conn.Do("set","newsWeb",buffer.Bytes())

		fmt.Println("从mysql中获取数据")

	}else {
		//将从redis数据库中读出的数据反序列化
		//创建解码器
		dec := gob.NewDecoder(bytes.NewReader(result))
		//进行解码
		dec.Decode(&articleTypes)
		fmt.Println("从redis中获取数据")
	}

	//展示文章类型数据
	this.Data["articleTypes"] = articleTypes

	//传递数据给前台
	//展示文章数据
	this.Data["articles"] = articles
	//显示总记录数
	this.Data["count"] = count
	//显示总页数
	this.Data["pageCount"] = pageCount
	//显示当前页数
	this.Data["pageNum"] = pageNum

	//传递当下拉框选择的类型名给视图
	this.Data["typeName"]=typeName

	//渲染视图
	this.Layout="Layout.html"
	this.TplName = "index.html"

}

//展示添加文章页面
func (this *ArticleController) AddArticle() {
	//创建orm对象
	o := orm.NewOrm()

	//创建文章类别对象结构体切片
	var articleTypes []models.ArticleType

	//指定要查询的数据库
	o.QueryTable("ArticleType").All(&articleTypes)

	//传递数据到页面
	this.Data["articleTypes"] = articleTypes

	//返回数据
	this.Layout="Layout.html"
	this.TplName = "add.html"
}

//添加文章业务处理
func (this *ArticleController) HandleAddArticle() {
	articleName := this.GetString("articleName")
	articleContent := this.GetString("content")
	typeName := this.GetString("select")

	//校验数据
	if articleName == "" || articleContent == "" || typeName == "" {
		fmt.Println("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//获取图片
	imgFile, fileHeader, err := this.GetFile("uploadname")

	defer imgFile.Close()
	//图片校验,错误
	if err != nil {
		fmt.Println("图片获取错误")
		this.Data["errmsg"] = "图片获取错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//校验图片大小
	if fileHeader.Size > 5000000 {
		fmt.Println("图片大小错误")
		this.Data["errmsg"] = "图片大小错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//校验图片格式
	ext := path.Ext(fileHeader.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		fmt.Println("图片格式错误")
		this.Data["errmsg"] = "图片格式错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//防止重名
	//获取新名
	fileName := time.Now().Format("200601021504052222")

	err = this.SaveToFile("uploadname", "./static/img/"+fileName+ext)
	if err != nil {
		fmt.Println("图片保存错误")
		this.Data["errmsg"] = "图片保存错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//存储数据
	o := orm.NewOrm()

	var article models.Article

	article.Title = articleName
	article.Content = articleContent
	article.Img = "/static/img/" + fileName + ext

	//获取一个文章类型对象
	var articleType models.ArticleType
	articleType.Typename = typeName
	//如果查询时不是主建查询，必须加查询条件
	o.Read(&articleType,"Typename")

	//将文章类型赋值给文章结构体中的文章类型
	article.ArticleType = &articleType

	_, err = o.Insert(&article)
	if err != nil {
		fmt.Println("数据插入错误")
		this.Data["errmsg"] = "数据插入错误"
		this.Layout="Layout.html"
		this.TplName = "add.html"
		return
	}

	//跳转首页
	this.Redirect("/article/index", 302)

}

//查看文章详情【重点】
func (this *ArticleController) ShowContent() {
	id, err := this.GetInt("id")
	if err != nil {
		fmt.Println("ID请求错误")
		this.Redirect("/article/index", 302)
		return
	}

	o := orm.NewOrm()

	var article models.Article

	article.ID = id

	o.Read(&article)

	//-------------多对多查询-------------------
	//多对多查询一
	//o.LoadRelated(&article,"Peoples")

	//多对多查询二
	//定义结构体切片，用来储存擦查到的内容
	var users []models.People
	//高级查询                  要查询的表名       字段名   字段对应的类型名  字段名  要查询的值       查询所有  查询一个用one()
	o.QueryTable("People").Filter("Articles__Article__ID",id).Distinct().All(&users)
    this.Data["users"]=users

	//增加阅读次数
	article.ReadConst += 1
	o.Update(&article)

//------------------------------------------
	//获取查看文章用户的用户名
	userName := this.GetSession("userName")

	//插入用户与文章关系数据-----多对多
	//获取映射对象--前段代码定义过了
	//获取被插入对象--前段代码定义过了，不能重新定义，就是要对正在查阅的文章对应的对象的数据进行操作，插入用户与文章的对应关系
	//获取要插入对象
	var user models.People
	//给查询对象赋值
	user.Name=userName.(string)
	//获取要插入对象全部信息
	o.Read(&user,"Name")

	//获取多对多插入对象
	m2m := o.QueryM2M(&article,"Peoples")

	//用多对多插入对象插入数据
	m2m.Add(user)

//-------------------------------------------------------
	this.Data["article"] = article
	//添加视图模板
	this.Layout="Layout.html"
	this.TplName = "content.html"
}

//展示编辑文章页面
func (this *ArticleController) ShowEdit() {
	//获取数据
	id, err := this.GetInt("id")

	//校验数据
	if err != nil {
		fmt.Println("查看编辑页面错误")
		this.Redirect("/article/index", 302)
		return
	}

	//处理数据
	o := orm.NewOrm()

	var article models.Article

	article.ID = id

	err = o.Read(&article)
	if err != nil {
		fmt.Println("查询数据错误")
		this.Redirect("/article/index", 302)
		return
	}

	//返回数据
	this.Data["article"] = article

	//添加视图模板
	this.Layout="Layout.html"
	this.TplName = "update.html"

}

//更新文章数据
func (this *ArticleController) UpdateArticle() {
	//获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	savepath := upImgPath(this, "uploadname", "/edit")
	id, _ := this.GetInt("id")

	//校验数据
	if articleName == "" || content == "" || savepath == "" {
		fmt.Println("获取数据失败")
		this.Redirect("/article/edit?id=0"+strconv.Itoa(id), 302)
		return
	}

	//处理数据
	o := orm.NewOrm()
	var article models.Article
	article.ID = id
	o.Read(&article)
	//更新数据
	article.Title = articleName
	article.Content = content
	article.Img = savepath
	o.Update(&article)

	//返回数据
	this.Redirect("/article/index", 302)
}

//进行图片检验
func upImgPath(this *ArticleController, filePath string, errhtml string) string {
	//获取图片
	imgFile, fileHeader, err := this.GetFile(filePath)

	defer imgFile.Close()
	//图片校验,错误
	if err != nil {
		fmt.Println("图片获取错误")
		this.Data["errmsg"] = "图片获取错误"
		this.TplName = errhtml
		return ""
	}

	//校验图片大小
	if fileHeader.Size > 5000000 {
		fmt.Println("图片大小错误")
		this.Data["errmsg"] = "图片大小错误"
		this.TplName = errhtml
		return ""
	}

	//校验图片格式
	ext := path.Ext(fileHeader.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		fmt.Println("图片格式错误")
		this.Data["errmsg"] = "图片格式错误"
		this.TplName = errhtml
		return ""
	}

	//防止重名
	//获取新名
	fileName := time.Now().Format("200601021504052222")

	err = this.SaveToFile(filePath, "./static/img/"+fileName+ext)
	if err != nil {
		fmt.Println("图片保存错误")
		this.Data["errmsg"] = "图片保存错误"
		this.TplName = errhtml
		return ""
	}
	return "/static/img/" + fileName + ext
}

//删除文章
func (this *ArticleController) HandleDelete() {
	//获取数据
	id, err := this.GetInt("id")

	//校验数据
	if err != nil {
		fmt.Print("获取数据失败")
		this.Redirect("/article/index", 302)
		return
	}

	//处理数据
	o := orm.NewOrm()
	var article models.Article
	article.ID = id
	o.Delete(&article)

	//返回数据
	this.Redirect("/article/index", 302)
}

//展示添加文章分类页面
func (this *ArticleController) ShowArticleType() {
	//创建映射对象
	o := orm.NewOrm()

	//创建文章分类对象切片
	var ArticleTypes []models.ArticleType

	//指定要查询的数据库，进行高级查询
	o.QueryTable("ArticleType").All(&ArticleTypes)

	//返回数据
	this.Data["ArticleTypes"] = ArticleTypes
	this.Layout="Layout.html"
	this.TplName = "addType.html"
}

//操作添加类型
func (this *ArticleController) HandleAddType() {
	//获取数据
	typeName := this.GetString("typeName")

	//检验数据
	if typeName == "" {
		fmt.Println("获取数据失败,类型添加失败")
		this.Redirect("/article/addType", 302)
		return
	}

	//处理数据
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Typename = typeName
	_, err := o.Insert(&articleType)
	if err != nil {
		fmt.Println("添加类型失败")
		this.Redirect("/article/addType", 302)
		return
	}

	//反获数据
	this.Redirect("/article/addType", 302)
}

//删除文章类型
func (this *ArticleController)DeleteType(){
	//获取数据
    id,err:=this.GetInt("Id")

	//校验数据
	if err!=nil{
		fmt.Println("获取文章id失败")
		this.Redirect("/article/addType",302)
		return
	}

	//处理数据
	o:=orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id=id
	o.Delete(&articleType)

	//返回数据
	this.Redirect("/article/addType",302)
}