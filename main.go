package main

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// main ...
func main() {
	// 初始化WEB服务
	r := gin.Default()
	r.Use(gin.ErrorLogger())

	// 连接数据库
	connStr := os.Getenv("DBUser") + ":" + os.Getenv("DBPassword") +
		"@(" + os.Getenv("DBHost") + ":" + os.Getenv("DBPort") + ")/" + os.Getenv("DBName")
	db := sqlx.MustConnect("mysql", connStr)
	r.Use(database(db))

	// session保持在cookie里
	store := cookie.NewStore([]byte("pfl+1535325623"))
	r.Use(sessions.Sessions("session", store))

	// 静态文件
	r.Use(static.Serve("/", static.LocalFile("./static", false)))
	r.StaticFile("/", "./static/index.html")

	// 用户
	r.POST("/api/user", register)       // 用户注册
	r.POST("/api/session", login)       // 用户登录
	r.GET("/api/logout", logout)        // 注销登录
	r.PUT("/api/user", updatePassword)  // 更改密码
	r.PUT("/api/modPhone", updatePhone) // 更改手机号

	// 二进制
	r.POST("/api/image", uploadImage)       // 上传图片
	r.GET("/api/image/:id", showImage)      // 查看图片
	r.DELETE("/api/image/:id", deleteImage) // 删除图片

	// 验证码
	r.GET("/api/verify", getVerifyCode)   // 获取验证码
	r.GET("/api/sendVerify", sendVerify)  // 发送验证码
	r.POST("/api/setDysms", setDysms)     // 设置短信
	r.POST("/api/setMail", setMail)       // 设置邮箱
	r.GET("/api/dysmsList", dysmsList)    // 短信模板列表
	r.GET("/api/mailList", mailList)      // 邮箱模板列表
	r.DELETE("/api/delTemp", delTemplate) // 删除短信/邮箱模板

	// 身份认证
	r.POST("/api/validate", createValidate)       // 提交身份认证
	r.GET("/api/validate", listValidate)          // 身份认证列表
	r.GET("/api/show_validate", showValidate)     // 查看身份认证
	r.PUT("api/validate", checkValidate)          // 审核身份认证
	r.DELETE("/api/validate/:id", deleteValidate) // 删除身份认证

	// 宝贝上传
	r.POST("/api/treasure", createTreasure)    // 上传宝贝
	r.GET("/api/treasure", listTreasure)       // 宝贝列表
	r.GET("/api/treasure/:id", showTreasure)   // 宝贝详情
	r.DELETE("/api/treasure", deleteTreasure)  // 删除宝贝
	r.PUT("/api/treasure", checkTreasure)      // 审核宝贝
	r.PUT("/api/treasure/:id", updateTreasure) // 更新宝贝

	// 宝贝交易
	r.POST("/api/order", listOrder)     // 订单列表
	r.DELETE("/api/order", deleteOrder) // 删除订单

	// 委托
	r.POST("/api/consign", creatConsign)        // 创建委托
	r.GET("/api/consign", listConsign)          // 委托列表
	r.PUT("/api/consign", putConsign)           // 撤回委托
	r.DELETE("/api/consign/:id", deleteConsign) // 删除委托

	// 资产
	r.GET("/api/assets", listAssets)          // 我的资产
	r.DELETE("/api/assets/:id", deleteAssets) // 删除资产

	// K线
	r.GET("/api/KLine/:id", KLine) // K线
	r.Run(":8080")
}
