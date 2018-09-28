package main

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"regexp"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var (
	timeFormat = "2006-01-02 15:04:05"
)

// register 用户注册
func register(c *gin.Context) {
	type User struct {
		Username      string `db:"username"       json:"username"      binding:"required"`
		Password      string `db:"password"       json:"password"      binding:"required"`
		MailBox       string `db:"mailbox"        json:"mailbox"       binding:"required"`
		FundsPassword string `db:"funds_password" json:"fundsPassword" binding:"required"`
		InviteCode    string `db:"invite_code"    json:"inviteCode,omitempty"`
		CreateTime    string `db:"create_time"`
		Code          string `                    json:"code"`
	}
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "请求参数有误"})
		return
	}
	// 验证码验证
	session := sessions.Default(c)
	verify, ok := session.Get("code").(string)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "验证码已失效"})
		return
	}
	if user.Code != verify {
		c.JSON(200, gin.H{"code": 401, "msg": "验证码验证失败"})
		return
	}
	// 删除验证码
	session.Delete("code")
	session.Save()

	// md5加密密码
	sum := md5.Sum([]byte(user.Password))
	user.Password = hex.EncodeToString(sum[:])
	// md5交易密码
	sum = md5.Sum([]byte(user.FundsPassword))
	user.FundsPassword = hex.EncodeToString(sum[:])
	// 创建时间
	user.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	db := c.MustGet("DB").(*sqlx.DB)
	query := "SELECT EXISTS(SELECT id FROM users WHERE username = ? OR mailbox = ? OR phone = ?)"
	var exists bool
	db.QueryRow(query, user.Username, user.Username, user.Username).Scan(&exists)
	if exists {
		c.JSON(200, gin.H{"code": 401, "msg": "该账号已注册!"})
		return
	}
	query = "INSERT INTO users (username, password, mailbox, funds_password, invite_code, create_time, real_name, id_card) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(query, user.Username, user.Password, user.MailBox, user.FundsPassword, user.InviteCode, user.CreateTime, "", "")
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "注册失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "注册成功"})
}

// login 用户登录
func login(c *gin.Context) {
	type User struct {
		Username string `db:"username"  json:"username"         binding:"required"`
		Password string `db:"password"  json:"password,omitempty"`
		Code     string `               json:"code,omitempty"`
		BindIP   bool   `               json:"bindIp,omitempty"`
		Status   bool   `               json:"status,omitempty"`
	}
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "请求参数有误"})
		return
	}
	session := sessions.Default(c)
	db := c.MustGet("DB").(*sqlx.DB)
	// 如果密码为空
	if user.Password == "" {
		// 验证码验证
		verify, ok := session.Get("code").(string)
		if !ok {
			c.JSON(200, gin.H{"code": 401, "msg": "验证码已失效"})
			return
		}
		if user.Code != verify {
			c.JSON(200, gin.H{"code": 401, "msg": "验证码验证失败"})
			return
		}
		// 删除验证码
		session.Delete("code")
		session.Save()
		query := "SELECT id, username, grade, phone FROM users WHERE phone = ?"
		type Info struct {
			ID       int64  `db:"id"       json:"id"`
			Username string `db:"username" json:"username"`
			Grade    int    `db:"grade"    json:"grade"`
			Phone    string `db:"phone"    json:"phone"`
		}
		var info Info
		err = db.QueryRow(query, user.Username).Scan(&info.ID, &info.Username, &info.Grade, &info.Phone)
		if err != nil {
			//c.JSON(200, gin.H{"code": 401, "msg": "登录失败"})
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		// 存入session
		session.Set("uid", info.ID)
		err = session.Save()
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "登录失败"})
			return
		}
		//查询上次登录时间
		query = "SELECT MAX(create_time) FROM login_log WHERE user_id = ?"
		var lastTime string
		db.QueryRow(query, info.ID).Scan(&lastTime)
		//写入当前登录日志
		query = "INSERT INTO login_log(user_id, last_time, create_time, bind_ip) VALUES (?, ?, ?, ?)"
		createTime := time.Now().Format(timeFormat)
		var bindID string
		if user.BindIP {
			bindID = getIP()
		}
		db.Exec(query, info.ID, lastTime, createTime, bindID)
		c.JSON(200, gin.H{"code": 200, "data": info})
	} else { // 密码存在
		// md5加密密码
		sum := md5.Sum([]byte(user.Password))
		user.Password = hex.EncodeToString(sum[:])
		//  前后台登录
		if !user.Status { // 前台登录
			// 验证码验证
			if user.Code != "" {
				verify, ok := session.Get("code").(string)
				if !ok {
					c.JSON(200, gin.H{"code": 401, "msg": "验证码已失效"})
					return
				}
				if user.Code != verify {
					c.JSON(200, gin.H{"code": 401, "msg": "验证码验证失败"})
					return
				}
				// 删除验证码
				session.Delete("code")
				session.Save()
			}
			query := "SELECT id, username, grade, phone FROM users WHERE username = ? AND password = ?"
			type Info struct {
				ID       int64  `db:"id"       json:"id"`
				Username string `db:"username" json:"username"`
				Grade    int    `db:"grade"    json:"grade"`
				Phone    string `db:"phone"    json:"phone"`
			}
			var info Info
			err = db.QueryRow(query, user.Username, user.Password).Scan(&info.ID, &info.Username, &info.Grade, &info.Phone)
			if err != nil {
				c.JSON(200, gin.H{"code": 401, "msg": "登录失败"})
				return
			}
			// 存入session
			session.Set("uid", info.ID)
			err = session.Save()
			if err != nil {
				c.JSON(200, gin.H{"code": 401, "msg": "登录失败"})
				return
			}
			//查询上次登录时间
			query = "SELECT MAX(create_time) FROM login_log WHERE user_id = ?"
			var lastTime string
			db.QueryRow(query, info.ID).Scan(&lastTime)
			//写入当前登录日志
			query = "INSERT INTO login_log(user_id, last_time, create_time, bind_ip) VALUES (?, ?, ?, ?)"
			createTime := time.Now().Format(timeFormat)
			var bindID string
			if user.BindIP {
				bindID = getIP()
			}
			db.Exec(query, info.ID, lastTime, createTime, bindID)
			c.JSON(200, gin.H{"code": 200, "data": info})
		} else { // 后台登录
			query := "SELECT id, username, role FROM users WHERE username = ? AND password = ?"
			type Info struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			}
			var info Info
			var role int
			db.QueryRow(query, user.Username, user.Password).Scan(&info.ID, &info.Username, &role)
			if role != 1 {
				c.JSON(200, gin.H{"code": 401, "msg": "无访问权限"})
				return
			} else {
				// 存入session
				session = sessions.Default(c)
				session.Set("uid", info.ID)
				err = session.Save()
				if err != nil {
					c.JSON(200, gin.H{"code": 401, "msg": "登录失败"})
					return
				}
				c.JSON(200, gin.H{"code": 200, "data": info})
			}
		}
	}

}

// logout 注销登录
func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "注销成功"})
}

// getIP 获取IP地址
func getIP() (IP string) {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				IP = ipnet.IP.String()
			}
		}
	}
	return IP
}

// updatePassword 更新密码
func updatePassword(c *gin.Context) {
	type Pass struct {
		Username    string `json:"username,omitempty"`
		OldPassWord string `json:"oldPassword,omitempty"`
		NewPassWord string `json:"newPassword"`
		Code        string `json:"code"`
		Funds       bool   `json:"type,omitempty"`
	}
	var pass Pass
	err := c.BindJSON(&pass)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误!"})
		return
	}
	session := sessions.Default(c)
	// 验证验证码
	verify, ok := session.Get("code").(string)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "验证码已失效"})
		return
	}
	if pass.Code != verify {
		c.JSON(200, gin.H{"code": 401, "msg": "验证码验证失败"})
		return
	}
	// 删除验证码
	session.Delete("code")
	session.Save()
	db := c.MustGet("DB").(*sqlx.DB)
	var exists bool
	// 旧密码存在 则修改密码
	if pass.OldPassWord != "" {
		userID, ok := session.Get("uid").(int64)
		if !ok {
			c.JSON(200, gin.H{"code": 401, "msg": "请先登录!"})
			return
		}
		if pass.OldPassWord == pass.NewPassWord {
			c.JSON(200, gin.H{"code": 401, "msg": "新密码不能与原密码相同!"})
			return
		}
		// md5加密
		sum := md5.Sum([]byte(pass.NewPassWord))
		pass.NewPassWord = hex.EncodeToString(sum[:])
		// 判断数据库原密码是否与新密码相同
		query := "SELECT EXISTS(SELECT id FROM users WHERE id = ? AND password = ?)"
		db.QueryRow(query, userID, pass.NewPassWord).Scan(&exists)
		if exists {
			c.JSON(200, gin.H{"code": 401, "msg": "新密码不能与原密码相同!"})
			return
		}
		query = "UPDATE users SET "
		if !pass.Funds {
			query += "password = ? "
		} else {
			query += "funds_password = ? "
		}
		query += "WHERE id = ?"
		_, err = db.Exec(query, pass.NewPassWord, userID)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "密码更新失败!"})
			return
		}
	}
	// 旧密码不存在 则忘记密码
	if pass.Username != "" {
		// 排除管理员
		if pass.Username != "admin" {
			query := "SELECT EXISTS(SELECT id FROM users WHERE (username = ? OR mailbox = ? OR phone = ?) AND role <> 1)"
			db.QueryRow(query, pass.Username, pass.Username, pass.Username).Scan(&exists)
			if !exists {
				c.JSON(200, gin.H{"code": 401, "msg": "该账号不存在!"})
				return
			}
			query = "UPDATE users SET "
			if !pass.Funds {
				query += "password = ? "
			} else {
				query += "funds_password = ? "
			}
			// 判断是账户是邮箱还是手机号或者用户名
			regMail := regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
			regPhone := regexp.MustCompile(`^1([38][0-9]|14[57]|5[^4])\d{8}$`)
			if regMail.MatchString(pass.Username) {
				// 如果是邮箱
				query += "WHERE mailbox = ? AND role <> 1"
			} else if regPhone.MatchString(pass.Username) {
				// 如果是手机号
				query += "WHERE phone = ? AND role <> 1"
			} else {
				// 如果是用户名
				query += "WHERE username = ? AND role <> 1"
			}
			_, err = db.Exec(query, pass.NewPassWord, pass.Username)
			if err != nil {
				c.JSON(200, gin.H{"code": 401, "msg": "密码修改失败!"})
				return
			}
			c.JSON(200, gin.H{"code": 200, "data": "密码修改成功!"})
			return
		} else {
			c.JSON(200, gin.H{"code": 401, "msg": "权限不足!"})
			return
		}
	}
	c.JSON(200, gin.H{"code": 401, "msg": "密码修改失败!"})
}

// updatePhone 更新手机号
func updatePhone(c *gin.Context) {
	type Pc struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var pc Pc
	err := c.BindJSON(&pc)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误!"})
		return
	}
	// 验证码验证
	session := sessions.Default(c)
	if pc.Code != "" {
		verify, ok := session.Get("code").(string)
		if !ok {
			c.JSON(200, gin.H{"code": 401, "msg": "验证码已失效"})
			return
		}
		if pc.Code != verify {
			c.JSON(200, gin.H{"code": 401, "msg": "验证码验证失败"})
			return
		}
		// 删除验证码
		session.Delete("code")
		session.Save()
	}
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请先登录!"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT id FROM users WHERE id = ? AND phone = ?)", userID, pc.Phone).Scan(&exists)
	if exists {
		c.JSON(200, gin.H{"code": 401, "msg": "手机号不能与原手机号码相同!"})
		return
	}
	_, err = db.Exec("UPDATE users SET phone = ? WHERE id = ?", pc.Phone, userID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "更新失败!"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "更新成功!", "phone": pc.Phone})
}
