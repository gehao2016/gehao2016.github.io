package main

import (
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Order struct {
	UserID      int64   `db:"user_id"      json:"userID"`
	TreasureID  int64   `db:"treasure_id"  json:"treasureID"`
	CreateTime  string  `db:"create_date"  json:"createDate"`
	OrderNumber string  `db:"order_number" json:"orderNumber"`
	ImageID     int64   `db:"image_id"     json:"imageID"`
	Type        int     `db:"type"         json:"type"`
	Price       float64 `db:"price"        json:"price"`
	Amount      float64 `db:"amount"       json:"amount"`
	Total       float64 `db:"total"        json:"total"`
	RealPay     float64 `db:"real_pay"     json:"realPay"`
	Status      int     `db:"status"       json:"status"`
	PriceFloat  float64 `db:"price_float"  json:"priceFloat"`
}

// listOrder 订单列表
func listOrder(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	type OrderList struct {
		ID          int64   `db:"id"           json:"id"`
		TreasureID  int64   `db:"treasure_id"  json:"treasureID"`
		CreateTime  string  `db:"create_date"  json:"createDate"`
		OrderNumber string  `db:"order_number" json:"orderNumber"`
		ImageID     int64   `db:"image_id"     json:"imageID"`
		Type        int     `db:"type"         json:"type"`
		Price       float64 `db:"price"        json:"price"`
		Amount      float64 `db:"amount"       json:"amount"`
		RealPay     float64 `db:"real_pay"     json:"realPay"`
		Status      int     `db:"status"       json:"status"`
		Title       string  `db:"title"        json:"title"`
		Roomnum     int     `db:"roomnum"      json:"roomnum"`
		Officenum   int     `db:"officenum"    json:"officenum"`
		Toiletnum   int     `db:"toiletnum"    json:"toiletnum"`
		Orientation int     `db:"orientation"  json:"orientation"`
		Decoration  int     `db:"decoration"   json:"decoration"`
		HouseType   int     `db:"house_type"   json:"houseType"`
		Acreage     float64 `db:"acreage"      json:"acreage"`
		Height      float64 `db:"height"       json:"height"`
		Floor       int     `db:"floor"        json:"floor"`
		Adress      string  `db:"adress"       json:"adress"`
		PriceFloat  float64 `db:"price_float"  json:"priceFloat"`
	}
	db := c.MustGet("DB").(*sqlx.DB)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	type Search struct {
		Page        uint   `json:"page,omitempty"`
		Type        uint   `json:"type,omitempty"`
		Lenth       uint   `json:"lenth,omitempty"`
		Title       string `json:"title,omitempty"`
		OrderNumber string `json:"orderNumber,omitempty"`
		TradeTime   string `json:"tradeTime,omitempty"`
		Evaluate    uint   `json:"evaluate,omitempty"`
		Status      uint   `json:"status,omitempty"`
	}
	var s Search
	err := c.BindJSON(&s)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误"})
		return
	}
	if s.Page < 1 {
		s.Page = 1
	}
	if s.Lenth == 0 {
		s.Lenth = 10
	}
	var args []interface{}
	query := "SELECT o.id, o.treasure_id, o.create_date, o.order_number, o.image_id, o.type, o.price, o.amount, " +
		"o.real_pay, o.`status`, o.price_float, h.roomnum, h.officenum, h.toiletnum, h.orientation,  " +
		"h.decoration, h.type AS house_type, h.acreage, h.height, h.floor, h.adress, t.title FROM orders AS o  " +
		"LEFT JOIN treasure AS t ON o.treasure_id = t.id LEFT JOIN house AS h ON t.house_id  = h.id WHERE o.user_id = ? "
	sql := "SELECT COUNT(o.id) FROM orders AS o LEFT JOIN treasure AS t ON o.treasure_id = t.id LEFT JOIN house AS h ON t.house_id  = h.id WHERE o.user_id = ? "
	args = append(args, userID)
	if s.Type != 0 {
		query += "AND o.type = ? "
		args = append(args, s.Type)
		sql += "AND o.type = ? "
	}
	if s.Title != "" {
		query += "AND t.title like ? "
		args = append(args, s.Title+"%")
		sql += "AND t.title like ? "
	}
	if s.OrderNumber != "" {
		query += "AND o.order_number like ? "
		args = append(args, s.OrderNumber+"%")
		sql += "AND o.order_number like ? "
	}
	if s.TradeTime != "" {
		arr := strings.Split(s.TradeTime, " ")
		query += "AND o.create_date between ? AND ? "
		args = append(args, arr[0], arr[1])
		sql += "AND o.create_date between ? AND ? "
	}
	if s.Evaluate != 0 {
		query += "AND o.evaluate = ? "
		args = append(args, s.Evaluate)
		sql += "AND o.evaluate = ? "
	}
	if s.Status != 0 {
		query += "AND o.status = ? "
		args = append(args, s.Status)
		sql += "AND o.status = ? "
	}
	//总条数
	var total int64
	db.QueryRow(sql, args...).Scan(&total)
	query += "ORDER BY o.create_time desc LIMIT ?, ? "
	args = append(args, (s.Page-1)*s.Lenth, s.Lenth)

	rows, err := db.Queryx(query, args...)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "暂无数据"})
		return
	}
	defer rows.Close()
	lists := []OrderList{}
	for rows.Next() {
		var list OrderList
		err = rows.StructScan(&list)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		lists = append(lists, list)
	}
	c.JSON(200, gin.H{"code": 200, "data": lists, "total": total})
}

// deleteOrder 删除订单
func deleteOrder(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	_, err := db.Exec("DELETE FROM orders WHERE user_id = ? AND id IN(?)", userID, c.Query("id"))
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "删除成功"})
}
