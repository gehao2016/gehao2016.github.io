package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Treasure struct {
	ID             int64   `db:"id"             json:"id"`
	HouseId        int64   `db:"house_id"       json:"houseId"`
	Sellway        int     `db:"sellway"        json:"sellway"`
	Openbusiness   string  `db:"openbusiness"   json:"openbusiness"`
	Opentime       string  `db:"opentime"       json:"opentime"`
	Housecard      string  `db:"housecard"      json:"housecard"`
	Housetype      int     `db:"housetype"      json:"housetype"`
	Propertyright  int     `db:"propertyright"  json:"propertyright"`
	Oneprice       float64 `db:"oneprice"       json:"oneprice,omitempty"`
	Decorationfree float64 `db:"decorationfree" json:"decorationfree"`
	Title          string  `db:"title"          json:"title"`
	Describe       string  `db:"describe"       json:"describe"`
	Contacts       string  `db:"contacts"       json:"contacts"`
	Phone          string  `db:"phone"          json:"phone"`
	Status         int     `db:"status"         json:"status"`
	Type           int     `db:"type"           json:"type"`
	Contract       string  `db:"contract"       json:"contract,omitempty"`
	Subtitle       string  `db:"subtitle"       json:"subtitle"`
	CreateTime     string  `db:"create_time"    json:"createTime"`
}

type House struct {
	ID          int64   `db:"id"           json:"id"`
	Roomnum     int     `db:"roomnum"      json:"roomnum"`
	Officenum   int     `db:"officenum"    json:"officenum"`
	Toiletnum   int     `db:"toiletnum"    json:"toiletnum"`
	Orientation int     `db:"orientation"  json:"orientation"`
	Decoration  int     `db:"decoration"   json:"decoration"`
	Type        int     `db:"type"         json:"type"`
	Acreage     float64 `db:"acreage"      json:"acreage"`
	Height      float64 `db:"height"       json:"height"`
	Floor       int     `db:"floor"        json:"floor"`
	Adress      string  `db:"adress"       json:"adress"`
	Evaluate    float64 `db:"evaluate"     json:"evaluate"`
	Houseimg    string  `db:"houseimg"     json:"houseimg"`
	Images      []int64 `                  json:"images"`
}

type HouseImage struct {
	ID      int64 `db:"id"`
	HouseID int64 `db:"house_id"`
	ImageID int64 `db:"image_id"`
}

// createTreasure创建宝贝
func createTreasure(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	type TreasureData struct {
		Treasure
		House `json:"houseDate"`
	}
	var data TreasureData
	err := c.BindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	db := c.MustGet("DB").(*sqlx.DB)
	// 开启事物
	tx, err := db.Begin()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	// 事物回滚
	defer tx.Rollback()
	//防止重复上传
	var exist bool
	tx.QueryRow("SELECT EXISTS(SELECT id FROM validate WHERE user_id = ? AND status = ?)", userID, 1).Scan(&exist)
	if !exist {
		c.JSON(200, gin.H{"code": 401, "msg": "请先身份认证"})
		return
	}
	tx.QueryRow("SELECT EXISTS(SELECT id FROM treasure WHERE title = ?)", data.Treasure.Title).Scan(&exist)
	if exist {
		c.JSON(200, gin.H{"code": 401, "msg": "请勿重复上传"})
		return
	}
	// 省市区
	pct := strings.Split(data.House.Adress, " ")
	// 房子
	query := "INSERT INTO house VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	res, err := tx.Exec(query, nil, data.House.Roomnum, data.House.Officenum, data.House.Toiletnum,
		data.House.Orientation, data.House.Decoration, data.House.Type, data.House.Acreage,
		data.House.Height, data.House.Floor, pct[3], data.House.Evaluate, pct[0], pct[1], pct[2], data.House.Houseimg)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "房屋信息录入失败"})
		return
	}
	houeseID, err := res.LastInsertId()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "获取房屋ID失败"})
		return
	}
	// 图片
	query = "INSERT INTO house_image (house_id, image_id) VALUES (?, ?)"
	for _, imageID := range data.House.Images {
		_, err = tx.Exec(query, houeseID, imageID)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "关联图片失败"})
			return
		}
	}
	data.Treasure.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	// 宝贝
	query = "INSERT INTO treasure VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	res, err = tx.Exec(query, nil, userID, houeseID, data.Treasure.Sellway, data.Treasure.Openbusiness,
		data.Treasure.Opentime, data.Treasure.Housecard, data.Treasure.Housetype, data.Treasure.Propertyright,
		data.Treasure.Oneprice, data.Treasure.Decorationfree, data.Treasure.Title, data.Treasure.Describe,
		data.Treasure.Contacts, data.Treasure.Phone, 3, data.Treasure.Type, data.Treasure.Contract,
		data.Treasure.Subtitle, data.Treasure.CreateTime)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "新增宝贝失败"})
		return
	}
	treasureID, err := res.LastInsertId()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "获取宝贝ID失败"})
		return
	}
	// 宝贝资产
	query = "INSERT INTO user_treasure VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(query, nil, userID, treasureID, data.House.Acreage)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "创建用户宝贝资产"})
		return
	}
	// 提交事物
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "上传成功"})
}

// listTreasure 宝贝列表
func listTreasure(c *gin.Context) {
	type List struct {
		ID         int64   `db:"id"          json:"id"`
		Evaluate   float64 `db:"evaluate"    json:"evaluate"`
		Title      string  `db:"title"       json:"title"`
		Acreage    float64 `db:"acreage"     json:"acreage"`
		Remain     float64 `db:"remain"      json:"remain"`
		Adress     string  `db:"adress"      json:"adress"`
		ImageId    int64   `db:"image_id"    json:"imageId"`
		Status     int     `db:"status"      json:"status"`
		Province   string  `db:"province"    json:"province"`
		City       string  `db:"city"        json:"city"`
		Area       string  `db:"area"        json:"area"`
		CreateTime string  `db:"create_time" json:"createTime"`
		PriceFloat float32 `                 json:"priceFloat"`
	}
	db := c.MustGet("DB").(*sqlx.DB)

	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "10"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	var query string
	var total int64
	var args []interface{}

	if c.Query("owner") == "true" {
		// 个人
		userID := sessions.Default(c).Get("uid").(int64)
		query = "SELECT t.id, t.title, h.evaluate, h.acreage, ut.remain, t.`status`, h.adress, hi.image_id, " +
			"t.create_time, h.province, h.city, h.area FROM treasure AS t LEFT JOIN user_treasure AS ut ON t.id = ut." +
			"treasure_id LEFT JOIN house AS h ON t.house_id = h.id LEFT JOIN house_image AS hi ON t.house_id = hi." +
			"house_id WHERE t.user_id = ? "
		args = append(args, userID)
		if c.Query("status") != "" {
			query += "AND t.status IN( " + c.Query("status") + " ) "
		}
		if c.Query("province") != "" {
			query += "AND h.province like ? "
			args = append(args, c.Query("province")+"%")
		}
		if c.Query("city") != "" {
			query += "AND h.city like ? "
			args = append(args, c.Query("city")+"%")
		}
		if c.Query("area") != "" {
			query += "AND h.area like ? "
			args = append(args, c.Query("area")+"%")
		}
		if c.Query("search") != "" {
			query += "AND t.title like ? OR h.adress like ? "
			args = append(args, c.Query("search")+"%", c.Query("search")+"%")
		}
		query += "GROUP BY t.id ORDER BY t.id DESC, t.status DESC LIMIT ?, ?"
		args = append(args, (page-1)*offset, offset)
		db.QueryRow("SELECT COUNT(t.id) FROM treasure AS t LEFT JOIN user_treasure AS ut ON t.id = ut.treasure_id WHERE ut.user_id = ?", userID).Scan(&total)
	} else {
		// 全体
		query = "SELECT t.id, t.title, h.evaluate, h.acreage, ut.remain, t.`status`, h.adress, hi.image_id, " +
			"t.create_time, h.province, h.city, h.area FROM treasure AS t LEFT JOIN user_treasure AS ut ON t.id = ut." +
			"treasure_id LEFT JOIN house AS h ON t.house_id = h.id LEFT JOIN house_image AS hi ON t.house_id = hi." +
			"house_id WHERE "
		if c.Query("status") != "" {
			query += "t.status IN( " + c.Query("status") + " ) "
		} else {
			query += "t.status = 1 "
		}
		if c.Query("province") != "" {
			query += "AND h.province like ? "
			args = append(args, c.Query("province")+"%")
		}
		if c.Query("city") != "" {
			query += "AND h.city like ? "
			args = append(args, c.Query("city")+"%")
		}
		if c.Query("area") != "" {
			query += "AND h.area like ? "
			args = append(args, c.Query("area")+"%")
		}
		if c.Query("search") != "" {
			query += "AND t.title like ? OR h.adress like ? "
			args = append(args, c.Query("search")+"%", c.Query("search")+"%")
		}
		query += "GROUP BY t.id ORDER BY t.id DESC, t.status DESC LIMIT ?, ?"
		args = append(args, (page-1)*offset, offset)
		db.QueryRow("SELECT COUNT(id) FROM treasure").Scan(&total)
	}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	var lists []List
	for rows.Next() {
		var list List
		err = rows.StructScan(&list)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		query = "SELECT price_float FROM orders WHERE treasure_id = ? AND status = 1 ORDER BY create_date DESC, " +
			"create_time DESC limit 1"
		db.QueryRow(query, list.ID).Scan(&list.PriceFloat)
		lists = append(lists, list)
	}
	c.JSON(200, gin.H{"code": 200, "data": lists, "total": total})
}

// showTreasure 宝贝详情
func showTreasure(c *gin.Context) {
	type ShowTreasure struct {
		Treasure
		Remain float64 `db:"remain"      json:"remain"`
		Status int     `db:"status"      json:"status"`
		House  `                         json:"houseDate"`
	}
	var show ShowTreasure
	show.Treasure.ID, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	db := c.MustGet("DB").(*sqlx.DB)
	query := "SELECT t.id, t.sellway, t.openbusiness, t.opentime, t.house_id, t.housecard, t.housetype, t.propertyright, " +
		"t.oneprice, t.decorationfree, t.title, t.`describe`, t.contacts, t.phone, t.type, t.contract, t.subtitle, ut.remain, h.evaluate, " +
		"t.`status`, h.roomnum, h.officenum, h.toiletnum, h.orientation, h.decoration, h.type, h.houseimg, " +
		"h.acreage, h.height, h.floor, h.adress FROM treasure AS t LEFT JOIN user_treasure AS ut ON t.id = ut.treasure_id " +
		"LEFT JOIN house AS h ON t.house_id = h.id WHERE t.id = ?"
	err := db.QueryRow(query, show.Treasure.ID).Scan(&show.Treasure.ID, &show.Treasure.Sellway, &show.Treasure.Openbusiness,
		&show.Treasure.Opentime, &show.Treasure.HouseId, &show.Treasure.Housecard, &show.Treasure.Housetype,
		&show.Treasure.Propertyright, &show.Treasure.Oneprice, &show.Treasure.Decorationfree, &show.Treasure.Title,
		&show.Treasure.Describe, &show.Treasure.Contacts, &show.Treasure.Phone, &show.Treasure.Type, &show.Treasure.Contract, &show.Treasure.Subtitle,
		&show.Remain, &show.House.Evaluate, &show.Status, &show.House.Roomnum, &show.House.Officenum, &show.House.Toiletnum,
		&show.House.Orientation, &show.House.Decoration, &show.House.Type, &show.House.Houseimg, &show.House.Acreage, &show.House.Height,
		&show.House.Floor, &show.House.Adress)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	rows, _ := db.Queryx("SELECT image_id FROM house_image WHERE house_id = ?", show.Treasure.HouseId)
	for rows.Next() {
		var imageID int64
		err := rows.Scan(&imageID)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		show.House.Images = append(show.House.Images, imageID)
	}
	rows.Close()
	c.JSON(200, gin.H{"code": 200, "data": show})
}

// deleteTreasure 删除宝贝
func deleteTreasure(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新先登录"})
		return
	}
	id := c.Query("id")
	if id == "" {
		c.JSON(200, gin.H{"code": 401, "msg": "参数不正确"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	tx, err := db.Begin()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	defer tx.Rollback()
	// 判断宝贝是否被购买
	var exist bool
	query := "SELECT EXISTS(SELECT id FROM user_treasure WHERE treasure_id = ? AND user_id <> ?)"
	tx.QueryRow(query, id, userID).Scan(&exist)
	if exist {
		c.JSON(200, gin.H{"code": 401, "msg": "不能删除有交易记录的宝贝"})
		return
	}
	// 删除图片
	query = "DELETE FROM image WHERE id IN (SELECT image_id FROM house_image WHERE house_id IN" +
		"(SELECT house_id FROM treasure WHERE id IN(?)))"
	_, err = tx.Exec(query, id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "图片删除失败"})
		return
	}
	// 删除关系表
	query = "DELETE FROM house_image WHERE house_id IN(SELECT house_id FROM treasure WHERE id IN(?))"
	_, err = tx.Exec(query, id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "关系表删除失败"})
		return
	}
	// 删除房子
	query = "DELETE FROM house WHERE id IN(SELECT house_id FROM treasure WHERE id IN(?))"
	_, err = tx.Exec(query, id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "房子删除失败"})
		return
	}
	// 删除宝贝
	query = "DELETE FROM treasure WHERE id IN(?)"
	_, err = tx.Exec(query, id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "宝贝删除失败"})
		return
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "宝贝删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "删除成功"})
}

// deleteAssets 删除资产
func deleteAssets(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新先登录"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数不正确"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	query := "DELETE FROM user_treasure WHERE user_id = ? AND treasure_id = ? AND remain = 0"
	_, err = db.Exec(query, userID, id)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "请检查资产是否为零"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}

// listAssets 资产
func listAssets(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请先登录"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	query := "SELECT ut.id, ut.remain, CONCAT(h.province, h.city, h.area, h.adress) AS address, ut.treasure_id " +
		"FROM user_treasure AS ut LEFT JOIN treasure AS t ON ut.treasure_id = t.id " +
		"LEFT JOIN house AS h ON t.house_id = h.id WHERE ut.user_id = ?"
	type Assets struct {
		ID         int64   `db:"id"          json:"id"`
		TreasureID int64   `db:"treasure_id" json:"treasureID"`
		Remain     float64 `db:"remain"      json:"remian"`
		Address    string  `db:"address"     json:"address"`
		Price      float64 `                 json:"price"`
		Total      float64 `                 json:"total"`
	}
	rows, err := db.Queryx(query, userID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "暂无资产"})
		return
	}
	defer rows.Close()
	var lists []Assets
	for rows.Next() {
		var list Assets
		err = rows.StructScan(&list)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		// 获取宝贝最新成交价
		query = "SELECT price FROM orders WHERE treasure_id = ?"
		db.QueryRow(query, list.TreasureID).Scan(&list.Price)
		if list.Price == 0 {
			query = "SELECT h.evaluate FROM treasure AS t INNER JOIN house AS h ON t.house_id = h.id WHERE t.id = ?"
			db.QueryRow(query, list.TreasureID).Scan(&list.Price)
		}
		list.Total = list.Price * list.Remain
		lists = append(lists, list)
	}
	// 统计资产总估平方
	var square, RMB, squareCoin float64
	for _, v := range lists {
		square += v.Remain
		RMB += v.Total
		squareCoin += v.Total
	}
	// 获取我的账户资金
	var assets float64
	query = "SELECT assets FROM users WHERE id = ?"
	db.QueryRow(query, userID).Scan(&assets)
	// 当月交易量
	var volume float64
	query = "SELECT SUM(real_pay) FROM orders WHERE user_id = ? AND create_date like ?"
	db.QueryRow(query, userID, time.Now().Format("2006-01")+"%").Scan(&volume)
	c.JSON(200, gin.H{"code": 200, "data": lists, "square": square, "RMB": RMB, "squareCoin": squareCoin, "assets": assets, "volume": volume})
}

// checkTreasure 审核宝贝
func checkTreasure(c *gin.Context) {
	session := sessions.Default(c)
	_, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "尚未登录"})
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
	_, err = db.Exec("UPDATE treasure SET status = ? WHERE id = ?", vali.Status, vali.ID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "操作失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "操作成功"})
}

// updateTreasure 修改宝贝
func updateTreasure(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请重新登录"})
		return
	}
	type TreasureData struct {
		Treasure
		House `json:"houseDate"`
	}
	var data TreasureData
	err := c.BindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	// 开启事物
	tx, err := db.Begin()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	// 事物回滚
	defer tx.Rollback()
	//防止重复上传
	var exist bool
	tx.QueryRow("SELECT EXISTS(SELECT id FROM validate WHERE user_id = ? AND status = 1)", userID).Scan(&exist)
	if !exist {
		c.JSON(200, gin.H{"code": 401, "msg": "请先身份认证"})
		return
	}
	tx.QueryRow("SELECT EXISTS(SELECT id FROM user_treasure WHERE treasure_id = ? AND user_id <> ?)", data.Treasure.ID,
		userID).Scan(&exist)
	if exist {
		c.JSON(200, gin.H{"code": 401, "msg": "不能修改有交易记录的宝贝"})
		return
	}
	tx.QueryRow("SELECT EXISTS(SELECT id FROM treasure WHERE id = ? AND status <> 1 AND user_id = ?)", data.Treasure.ID,
		userID).Scan(&exist)
	if exist {
		c.JSON(200, gin.H{"code": 401, "msg": "不能修改已审核的宝贝"})
		return
	}
	// 获取宝贝ID
	var houseid int64
	tx.QueryRow("SELECT house_id FROM treasure WHERE id = ?", data.Treasure.ID).Scan(&houseid)
	// 更新宝贝信息
	query := "UPDATE FROM treasure SET sellway = ?, openbusiness = ?, opentime = ?, housecard = ?, housetype = ?," +
		"propertyright = ?, oneprice = ?, decoratinefree = ?, title = ?, describe = ?, contacts = ?, phone = ?, " +
		"status = 3, type = ?, contract = ?, subtitle = ? WHERE id = ?"
	_, err = tx.Exec(query, data.Treasure.Sellway, data.Treasure.Openbusiness, data.Treasure.Opentime,
		data.Treasure.Housecard,
		data.Treasure.Housetype, data.Treasure.Propertyright, data.Treasure.Oneprice, data.Treasure.Decorationfree,
		data.Treasure.Title, data.Treasure.Describe, data.Treasure.Contacts, data.Treasure.Phone, data.Treasure.Type,
		data.Treasure.Contract, data.Treasure.Subtitle, data.Treasure.ID)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "更新宝贝信息失败"})
		return
	}
	// 更新房屋信息
	query = "UPDATE FROM house SET roomnum = ?, offcenum = ?, toiletnum = ?, orientation = ?, decoration = ?, " +
		"type = ?, acreage = ?, height = ?, floor = ?, adress = ?, evaluate = ?, province = ?, city = ?, area = ?, " +
		"houseimg = ? WHERE id = ?"
	pct := strings.Split(data.House.Adress, " ")
	_, err = tx.Exec(query, data.House.Roomnum, data.House.Officenum, data.House.Toiletnum, data.House.Orientation,
		data.House.Decoration, data.House.Type, data.House.Acreage, data.House.Height, data.House.Floor, pct[3],
		data.House.Evaluate, pct[0], pct[1], pct[2], data.House.Houseimg, houseid)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "更新房屋信息失败"})
		return
	}
	if len(data.House.Images) != 0 {
		// 删除图片
		_, err = tx.Exec("DELETE FROM house_image WHERE house_id = ?", houseid)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "删除图片失败"})
			return
		}
		for _, image := range data.House.Images {
			_, err = tx.Exec("INSERT INTO house_image VALUES (?, ?, ?)", nil, houseid, image)
			if err != nil {
				c.JSON(200, gin.H{"code": 401, "msg": "添加图片失败"})
				return
			}
		}
	}
	// 提交事物
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "提交事物失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "更新成功"})
}
