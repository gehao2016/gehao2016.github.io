package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/smtp"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SendSmsReply 发送短信返回
type SendSmsReply struct {
	Code    string `json:"Code,omitempty"`
	Message string `json:"Message,omitempty"`
}

func replace(in string) string {
	rep := strings.NewReplacer("+", "%20", "*", "%2A", "%7E", "~")
	return rep.Replace(url.QueryEscape(in))
}

// getVerifyCode 获取6位验证码
func getVerifyCode(c *gin.Context) {
	code, err := verify(c)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "生成验证码失败!"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": code})
}

//  sendVerify 发送验证码
func sendVerify(c *gin.Context) {
	name := c.Query("name")
	switch name {
	case "mail": // 发送邮件
		code, err := verify(c)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "生成验证码失败!"})
			return
		}
		mail := c.Query("mail")
		type Mail struct {
			Username string `db:"username"`
			Password string `db:"password"`
			Nickname string `db:"nickname"`
			Subject  string `db:"subject"`
			Content  string `db:"content"`
		}
		var m Mail
		db := c.MustGet("DB").(*sqlx.DB)
		query := "SELECT username, password, nickname, subject, content FROM mail WHERE status = 1"
		err = db.QueryRow(query).Scan(&m.Username, &m.Password, &m.Nickname, &m.Subject, &m.Content)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "获取失败！"})
			return
		}
		if err = SendMail(m.Username, m.Password, mail, m.Nickname, m.Subject, code, m.Content); err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "发送失败！"})
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": "发送成功！"})
	case "sms": // 发送短信
		code, err := verify(c)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "生成验证码失败!"})
			return
		}
		db := c.MustGet("DB").(*sqlx.DB)
		query := "SELECT keyid, secret, signname, code FROM dysms WHERE status = 1 AND type = ?"
		type Dysms struct {
			accessKeyID   string
			accessSecret  string
			phoneNumbers  string
			signName      string
			templateParam string
			templateCode  string
		}
		var dysms Dysms
		err = db.QueryRow(query, c.DefaultQuery("type", "1")).Scan(&dysms.accessKeyID, &dysms.accessSecret,
			&dysms.signName,
			&dysms.templateCode)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "获取失败！"})
			return
		}
		dysms.templateParam = `{"code":"` + code + `"}`
		dysms.phoneNumbers = c.Query("phone")
		if err = SendSms(dysms.accessKeyID, dysms.accessSecret, dysms.phoneNumbers, dysms.signName,
			dysms.templateParam, dysms.templateCode); err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "发送失败！"})
		} else {
			c.JSON(200, gin.H{"code": 200, "data": "发送成功！"})
		}
	}
}

// verify 封装验证码
func verify(c *gin.Context) (code string, err error) {
	// 生成6位验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code = fmt.Sprintf("%06v", rnd.Int31n(1000000))

	// 保存在Session里
	session := sessions.Default(c)
	session.Set("code", code)
	err = session.Save()
	if err != nil {
		return "", errors.New("生成验证码失败!")
	}
	return code, nil
}

// SendSms 发送短信
func SendSms(accessKeyID, accessSecret, phoneNumbers, signName, templateParam, templateCode string) error {
	paras := map[string]string{
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", rand.Int63()),
		"AccessKeyId":      accessKeyID,
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Format":           "JSON",
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"RegionId":         "cn-hangzhou",
		"PhoneNumbers":     phoneNumbers,
		"SignName":         signName,
		"TemplateParam":    templateParam,
		"TemplateCode":     templateCode,
	}

	var keys []string

	for k := range paras {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var sortQueryString string

	for _, v := range keys {
		sortQueryString = fmt.Sprintf("%s&%s=%s", sortQueryString, replace(v), replace(paras[v]))
	}

	stringToSign := fmt.Sprintf("GET&%s&%s", replace("/"), replace(sortQueryString[1:]))

	mac := hmac.New(sha1.New, []byte(fmt.Sprintf("%s&", accessSecret)))
	mac.Write([]byte(stringToSign))
	sign := replace(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	str := fmt.Sprintf("http://dysmsapi.aliyuncs.com/?Signature=%s%s", sign, sortQueryString)

	resp, err := http.Get(str)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ssr := &SendSmsReply{}

	if err := json.Unmarshal(body, ssr); err != nil {
		return err
	}

	if ssr.Code == "SignatureNonceUsed" {
		return SendSms(accessKeyID, accessSecret, phoneNumbers, signName, templateParam, templateCode)
	} else if ssr.Code != "OK" {
		return errors.New(ssr.Code)
	}

	return nil
}

// SendMail 发送邮件
func SendMail(username, password, mail, nickname, subject, code, content string) error {
	// 身份认证
	auth := smtp.PlainAuth("", username, password, "smtp.qq.com")
	// 接收者
	to := strings.Split(mail, ",")
	// 文本编码方式
	content_type := "Content-Type: text/plain; charset=UTF-8"
	// 发送内容
	body := fmt.Sprintf(content, code)
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + username + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	// 发送邮件
	err := smtp.SendMail("smtp.qq.com:25", auth, username, to, msg)
	return err
}

// setDysms 设置短信
func setDysms(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.Error(errors.New("请先登录!")).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var exists bool
	query := "SELECT EXISTS(SELECT id FROM users WHERE id = ? AND role = 1)"
	db.QueryRow(query, userID).Scan(&exists)
	if !exists {
		c.Error(errors.New("无操作权限!")).SetType(gin.ErrorTypePrivate)
		return
	}
	type Dysms struct {
		ID       int64  `db:"id"       json:"id,omitempty"`
		KeyId    string `db:"keyid"    json:"keyid"`
		Secret   string `db:"secret"   json:"secret"`
		SignName string `db:"signname" json:"signname"`
		Code     string `db:"code"     json:"code"`
		Status   int    `db:"status"   json:"status"`
		Type     int    `db:"type"     json:"type"`
	}
	var dysms Dysms
	err := c.BindJSON(&dysms)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	// 有ID 则修改
	if dysms.ID != 0 {
		// 如果开启
		if dysms.Status == 1 {
			// 关闭其他同类
			_, err = tx.Exec("UPDATE dysms SET status = 2 WHERE status = 1 AND type = ?", dysms.Type)
			if err != nil {
				c.Error(errors.New("关闭失败!")).SetType(gin.ErrorTypePrivate)
				return
			}
		}
		// 更新短信模板
		query = "UPDATE dysms SET keyid = ?, secret = ?, signname = ?, code = ?, status = ?, type = ? WHERE id = ?"
		_, err = db.Exec(query, dysms.KeyId, dysms.Secret, dysms.SignName, dysms.Code, dysms.Status, dysms.Type, dysms.ID)
		if err != nil {
			c.Error(errors.New("更新失败!")).SetType(gin.ErrorTypePrivate)
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": "更新成功!"})
	} else {
		// 如果开启
		if dysms.Status == 1 {
			// 关闭其他同类
			_, err = tx.Exec("UPDATE dysms SET status = 2 WHERE status = 1 AND type = ?", dysms.Type)
			if err != nil {
				c.Error(errors.New("关闭失败!")).SetType(gin.ErrorTypePrivate)
				return
			}
		}
		// 新增短信模板
		query = "INSERT INTO dysms VALUES(?, ?, ?, ?, ?, ?, ?)"
		_, err = db.Exec(query, nil, dysms.KeyId, dysms.Secret, dysms.SignName, dysms.Code, dysms.Status, dysms.Type)
		if err != nil {
			c.Error(errors.New("新增失败!")).SetType(gin.ErrorTypePrivate)
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": "新增成功!"})
	}
}

// setMail 设置邮箱
func setMail(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.Error(errors.New("请先登录!")).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var exists bool
	query := "SELECT EXISTS(SELECT id FROM users WHERE id = ? AND role = 1)"
	db.QueryRow(query, userID).Scan(&exists)
	if !exists {
		c.Error(errors.New("无操作权限!")).SetType(gin.ErrorTypePrivate)
		return
	}
	type Mail struct {
		ID       int64  `db:"id"         json:"id,omitempty"`
		Username string `db:"username"   json:"username"`
		Password string `db:"password"   json:"password"`
		NickName string `db:"nickname"   json:"nickname"`
		Subject  string `db:"subject"    json:"subject" `
		Content  string `db:"content"    json:"content"`
		Status   int    `db:"status"     json:"status"`
	}
	var mail Mail
	err := c.BindJSON(&mail)
	if err != nil {
		c.Error(errors.New("参数有误!")).SetType(gin.ErrorTypePrivate)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	// 有无ID
	if mail.ID != 0 {
		// 开启
		if mail.Status == 1 {
			_, err = tx.Exec("UPDATE mail SET status = 2 WHERE status = 1")
			if err != nil {
				c.Error(err).SetType(gin.ErrorTypePrivate)
				return
			}
		}
		// 更新
		query = "UPDATE mail SET username = ?, password = ?, nickname = ?, subject = ?, content = ?, " +
			"status = ? WHERE id = ?"
		_, err = tx.Exec(query, mail.Username, mail.Password, mail.NickName, mail.Subject, mail.Content, mail.Status, mail.ID)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": "更新成功!"})
	} else {
		// 开启
		if mail.Status == 1 {
			_, err = tx.Exec("UPDATE mail SET status = 2 WHERE status = 1")
			if err != nil {
				c.Error(err).SetType(gin.ErrorTypePrivate)
				return
			}
		}
		// 新增
		query = "INSERT INTO mail VALUES(?, ?, ?, ?, ?, ?, ?)"
		_, err = tx.Exec(query, nil, mail.Username, mail.Password, mail.NickName, mail.Subject, mail.Content, mail.Status)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": "新增成功!"})
	}
}

// dysmsList 短信模板列表
func dysmsList(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.Error(errors.New("请先登录!")).SetType(gin.ErrorTypePrivate)
		return
	}
	type Dysms struct {
		ID       int64  `db:"id"       json:"id,omitempty"`
		KeyId    string `db:"keyid"    json:"keyid"`
		Secret   string `db:"secret"   json:"secret"`
		SignName string `db:"signname" json:"signname"`
		Code     string `db:"code"     json:"code"`
		Status   int    `db:"status"   json:"status"`
		Type     int    `db:"type"     json:"type"`
	}
	db := c.MustGet("DB").(*sqlx.DB)
	rows, err := db.Queryx("SELECT * FROM dysms ORDER BY status asc")
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	defer rows.Close()
	var list []Dysms
	for rows.Next() {
		var dysms Dysms
		err = rows.StructScan(&dysms)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		list = append(list, dysms)
	}
	c.JSON(200, gin.H{"code": 200, "data": list})
}

// mailList 邮箱模板列表
func mailList(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.Error(errors.New("请先登录!")).SetType(gin.ErrorTypePrivate)
		return
	}
	type Mail struct {
		ID       int64  `db:"id"         json:"id,omitempty"`
		Username string `db:"username"   json:"username"`
		Password string `db:"password"   json:"password"`
		NickName string `db:"nickname"   json:"nickname"`
		Subject  string `db:"subject"    json:"subject" `
		Content  string `db:"content"    json:"content"`
		Status   int    `db:"status"     json:"status"`
	}
	db := c.MustGet("DB").(*sqlx.DB)
	rows, err := db.Queryx("SELECT * FROM mail ORDER BY status asc")
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	defer rows.Close()
	var list []Mail
	for rows.Next() {
		var mail Mail
		err = rows.StructScan(&mail)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		list = append(list, mail)
	}
	c.JSON(200, gin.H{"code": 200, "data": list})
}

// delTemplate 删除短信/邮箱模板
func delTemplate(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.Error(errors.New("请先登录!")).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var exists bool
	query := "SELECT EXISTS(SELECT id FROM users WHERE id = ? AND role = 1)"
	db.QueryRow(query, userID).Scan(&exists)
	if !exists {
		c.Error(errors.New("无操作权限!")).SetType(gin.ErrorTypePrivate)
		return
	}
	name := c.Query("name")
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if name != "" && id != 0 {
		switch name {
		case "sms":
			_, err := db.Exec("DELETE FROM dysms WHERE id = ?", id)
			if err != nil {
				c.Error(err).SetType(gin.ErrorTypePrivate)
				return
			}
			c.JSON(200, gin.H{"code": 200, "data": "删除成功!"})
		case "mail":
			_, err := db.Exec("DELETE FROM mail WHERE id = ?", id)
			if err != nil {
				c.Error(err).SetType(gin.ErrorTypePrivate)
				return
			}
			c.JSON(200, gin.H{"code": 200, "data": "删除成功!"})
		}
	} else {
		c.Error(errors.New("参数有误!")).SetType(gin.ErrorTypePrivate)
		return
	}
}
