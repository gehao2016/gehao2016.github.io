package main

import (
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Auth struct {
	ID          int64  `db:"id"           json:"id"`
	RealName    string `db:"real_name"    json:"realName"`
	IdCard      string `db:"id_card"      json:"idCard"`
	Proct       string `db:"proct"        json:"proct"`
	Address     string `db:"address"      json:"address"`
	Sex         int    `db:"sex"          json:"sex"`
	Birthday    string `db:"birthday"     json:"birthday"`
	Qq          string `db:"qq"           json:"qq"`
	Phone       string `db:"phone"        json:"phone"`
	Spare       string `db:"spare"        json:"spare,omitempty"`
	HandAccount int64  `db:"hand_account" json:"handAccount"`
	Type        int    `db:"type"         json:"type"`
	Frontimg    int64  `db:"frontimg"     json:"frontimg"`
	Oppositeimg int64  `db:"oppositeimg"  json:"oppositeimg"`
	HandCard    int64  `db:"hand_card"    json:"handCard"`
	Status      int    `db:"status"       json:"status"`
}

// createValidate 提交身份信息
func createValidate(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	var auth Auth
	err := c.BindJSON(&auth)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "请求参数不正确"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	tx, err := db.Begin()
	if err != nil {
		c.JSON(200, gin.H{"code": 403, "msg": "禁止访问"})
		return
	}
	defer tx.Rollback()
	query := "UPDATE users SET real_name = ?, id_card = ?, proct = ?, address = ?, sex = ?, " +
		"birthday = ?, qq = ?, phone = ?, spare = ? WHERE id = ?"
	_, err = tx.Exec(query, auth.RealName, auth.IdCard, auth.Proct, auth.Address,
		auth.Sex, auth.Birthday, auth.Qq, auth.Phone, auth.Spare, userID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
		return
	}
	query = "SELECT EXISTS(SELECT id FROM validate WHERE user_id = ?)"
	var exists bool
	tx.QueryRow(query, userID).Scan(&exists)
	if !exists {
		query = "INSERT INTO validate (user_id, hand_account, type, frontimg, oppositeimg, hand_card) " +
			"VALUES (?, ?, ?, ?, ?, ?)"
		_, err = tx.Exec(query, userID, auth.HandAccount, auth.Type, auth.Frontimg, auth.Oppositeimg, auth.HandCard)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
			return
		}
	} else {
		query = "UPDATE validate SET hand_account = ?, type = ?, frontimg = ?, oppositeimg = ?, hand_card = ? " +
			"WHERE user_id = ?"
		_, err = tx.Exec(query, auth.HandAccount, auth.Type, auth.Frontimg, auth.Oppositeimg, auth.HandCard, userID)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "提交成功"})
}

// listValidate 身份认证列表
func listValidate(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var query string
	var total int64
	var args []interface{}
	if c.Query("type") != "" {
		typ, _ := strconv.ParseInt(c.DefaultQuery("type", "1"), 0, 0)
		query = "SELECT v.id, u.real_name, u.id_card, v.hand_account, v.type, v.frontimg, v.oppositeimg, v.hand_card, " +
			"v.status, u.address, u.birthday, u.phone, u.proct, u.qq, u.sex " +
			"FROM validate v INNER JOIN users u ON v.user_id = u.id WHERE v.type = ? ORDER BY v.status asc limit ?,10"
		args = append(args, typ, (page-1)*10)
		db.QueryRow("SELECT COUNT(v.id) FROM validate v INNER JOIN users u ON v.user_id = u.id WHERE v.type = ?", typ).Scan(&total)
	} else {
		query = "SELECT v.id, u.real_name, u.id_card, v.hand_account, v.type, v.frontimg, v.oppositeimg, v.hand_card, v.status, u.address, u.birthday, u.phone, u.proct, u.qq, u.sex " +
			"FROM validate AS v INNER JOIN users AS u ON v.user_id = u.id ORDER BY v.status ASC LIMIT ?, 10"
		args = append(args, (page-1)*10)
		db.QueryRow("SELECT COUNT(v.id) FROM validate v INNER JOIN users u ON v.user_id = u.id").Scan(&total)
	}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		c.JSON(200, gin.H{"code": 200, "msg": "暂无数据"})
		return
	}
	defer rows.Close()

	lists := []Auth{}
	for rows.Next() {
		var auth Auth
		err = rows.StructScan(&auth)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		lists = append(lists, auth)
	}
	c.JSON(200, gin.H{"code": 200, "data": lists, "total": total})
}

// showValidate 查看身份验证
func showValidate(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	var query string
	var args []interface{}
	if c.Query("id") != "" {
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		query = "SELECT v.id, u.real_name, u.id_card, v.hand_account, v.type, v.frontimg, v.oppositeimg, v.hand_card, v.status, u.address, u.birthday, u.phone, u.proct, u.qq, u.sex " +
			"FROM validate v INNER JOIN users u ON v.user_id = u.id WHERE v.id = ?"
		args = append(args, id)
	} else {
		query = "SELECT v.id, u.real_name, u.id_card, v.hand_account, v.type, v.frontimg, v.oppositeimg, v.hand_card, v.status, u.address, u.birthday, u.phone, u.proct, u.qq, u.sex " +
			"FROM validate v INNER JOIN users u ON v.user_id = u.id WHERE v.user_id = ?"
		args = append(args, userID)
	}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "暂无法查看"})
		return
	}
	defer rows.Close()
	var auth Auth
	for rows.Next() {
		err = rows.StructScan(&auth)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
	}
	c.JSON(200, gin.H{"code": 200, "data": auth})
}

// checkValidate 审核身份认证
func checkValidate(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	type Validate struct {
		ID     int64 `db:"id"     json:"id"`
		Status int   `db:"status" json:"status"`
	}
	var vali Validate
	err := c.BindJSON(&vali)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	_, err = db.Exec("UPDATE validate SET status = ? WHERE id = ?", vali.Status, vali.ID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "操作失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "操作成功"})
}

// deleteValidate 删除身份认证
func deleteValidate(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	_, err = db.Exec("DELETE FROM validate WHERE id = ?", id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "删除成功"})
}
