package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// 发布委托
func creatConsign(c *gin.Context) {
	type Consign struct {
		TreasureID int64   `db:"treasure_id" json:"treasureID"`
		CreateTime string  `db:"create_time" json:"createTime"`
		Type       int     `db:"type"        json:"type"`
		Price      float64 `db:"price"       json:"price"`
		Amount     float64 `db:"amount"      json:"amount"`
		Total      float64 `db:"total"       json:"total"`
	}
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "尚未登录"})
		return
	}
	var con Consign
	err := c.BindJSON(&con)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数不对"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	tx, err := db.Begin()
	if err != nil {
		c.JSON(200, gin.H{"code": 200, "msg": "事物开启失败"})
		return
	}
	defer tx.Rollback()
	var exist bool
	var query string
	if con.Type == 1 {
		//判断余额是否充足
		query = "SELECT EXISTS(SELECT assets FROM users WHERE id = ? AND assets > ?)"
		tx.QueryRow(query, userID, con.Price*con.Amount).Scan(&exist)
		if !exist {
			c.JSON(200, gin.H{"code": "401", "msg": "余额不足"})
			return
		}
	}
	if con.Type == 2 {
		// 判断可卖出量
		query = "SELECT EXISTS(SELECT id FROM user_treasure WHERE user_id = ? AND treasure_id = ? AND remain >= ?)"
		tx.QueryRow(query, userID, con.TreasureID, con.Amount).Scan(&exist)
		if !exist {
			c.JSON(200, gin.H{"code": "401", "msg": "可卖出数量不足"})
			return
		}
	}
	// 判断是否超出总面积
	query = "SELECT EXISTS(SELECT h.acreage FROM treasure AS t LEFT JOIN house AS h " +
		"ON t.house_id = h.id WHERE t.id = ? AND h.acreage > ?)"
	tx.QueryRow(query, con.TreasureID, con.Amount).Scan(&exist)
	if !exist {
		c.JSON(200, gin.H{"code": "401", "msg": "不能超出总面积"})
		return
	}
	// 判断是否存在相应资产
	query = "SELECT EXISTS(SELECT id FROM user_treasure WHERE user_id = ? AND treasure_id = ?)"
	tx.QueryRow(query, userID, con.TreasureID).Scan(&exist)
	// 不存在则创建
	if !exist {
		query = "INSERT INTO user_treasure VALUES (?, ?, ?, ?)"
		_, err = tx.Exec(query, nil, userID, con.TreasureID, 0)
		if err != nil {
			c.JSON(200, gin.H{"code": "401", "msg": "创建失败"})
			return
		}
	}
	//创建委托
	query = "INSERT INTO consign VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	con.CreateTime = time.Now().Format(timeFormat)
	res, err := tx.Exec(query, nil, userID, &con.TreasureID, &con.CreateTime, &con.Type,
		&con.Price, &con.Amount, 1, &con.Total, 0, 0)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "委托失败"})
		return
	}
	mcid, err := res.LastInsertId()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "获取委托ID失败"})
		return
	}
	// 订单
	var order Order
	// 随机数
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := fmt.Sprintf("%03v", rnd.Int31n(1000))
	// 订单号
	order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + randNum
	// 生成唯一订单号
Label:
	tx.QueryRow("SELECT EXISTS(SELECT id FROM orders WHERE order_number = ? )", order.OrderNumber).Scan(&exist)
	// 判断订单号是存在
	if exist {
		order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) +
			fmt.Sprintf("%03v", rnd.Int31n(1000))
		goto Label
	}
	// 实付款(包含手续费)
	order.RealPay = order.Total * (1 + 0.002)
	// 图片ID
	query = "SELECT MIN( image_id ) FROM house_image WHERE house_id = " +
		"(SELECT house_id FROM treasure WHERE id = ? )"
	tx.QueryRow(query, con.TreasureID).Scan(&order.ImageID)
	// 获取最新成交价
	query = "SELECT price FROM consign WHERE treasure_id = ? AND status = ? ORDER BY trade_time desc, id desc"
	var price float64
	tx.QueryRow(query, con.TreasureID, 2).Scan(&price)
	switch con.Type {
	case 1:
		// 冻结可用资金
		query = "UPDATE users SET assets = assets - ? WHERE id = ? AND assets >= ?"
		_, err = tx.Exec(query, con.Price*con.Amount, userID, con.Price*con.Amount)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "冻结可用资金失败"})
			return
		}
		// 查询匹配
		query = "SELECT id, user_id, (amount-trade_amount) AS num FROM consign " +
			"WHERE treasure_id = ? AND type = ? AND price <= ? AND amount - trade_amount > 0 ORDER BY price asc, " +
			"create_time asc"
		rows, err := db.Queryx(query, con.TreasureID, 2, con.Price)
		defer rows.Close()
		if err == nil {
			type Arr struct {
				Cid int64   `db:"id"`
				Uid int64   `db:"user_id"`
				Num float64 `db:"num"`
			}
			var arrs []Arr
			for rows.Next() {
				var arr Arr
				err = rows.StructScan(&arr)
				if err != nil {
					c.JSON(200, gin.H{"code": 401, "msg": "获取数据失败"})
					return
				}
				arrs = append(arrs, arr)
			}
			for _, v := range arrs {
				// 计算价格浮动
				if price > 0 {
					order.PriceFloat = (con.Price - price) / price
				} else {
					order.PriceFloat = 0
				}
				if con.Amount < v.Num { // 买入数量小于卖出数量
					// 买入
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, mcid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+1)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, con.Price, con.Amount, con.Price*con.Amount, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, con.Amount, userID, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = trade_amount + ?, trade_time = ? WHERE id = ?"
					_, err = tx.Exec(query, con.Amount, con.CreateTime, v.Cid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+1)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, con.Price, con.Amount, con.Price*con.Amount, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, con.Price*con.Amount, v.Uid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": " 增加卖方资金失败"})
						return
					}
				} else if con.Amount == v.Num { // 买入数量等于卖出数量
					// 买入
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, mcid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+3)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, con.Price, con.Amount, con.Price*con.Amount, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, con.Amount, userID, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ?  WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, v.Cid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+4)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, con.Price, v.Num, con.Price*v.Num, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, con.Price*con.Amount, v.Uid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": " 增加卖方资金失败"})
						return
					}
				} else { // 买入数量大于卖出数量
					// 买入
					query = "UPDATE consign SET trade_amount = trade_amount + ?, trade_time = ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, v.Num, con.CreateTime, userID, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+5)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, con.Price, v.Num, con.Price*v.Num, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, v.Num, userID, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, v.Cid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+6)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, con.Price, v.Num, con.Price*v.Num, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, con.Price*v.Num, v.Uid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "增加卖方资金失败"})
						return
					}
				}
				if con.Amount-v.Num > 0 {
					con.Amount = con.Amount - v.Num
				} else {
					// 扣除手续费
					query = "UPDATE users SET assets = assets - ? WHERE id = ? AND assets >= ?"
					_, err = tx.Exec(query, con.Total*0.002, userID, con.Total*0.002)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "扣除手续费失败"})
						return
					}
					// 收取手续费
					query = "UPDATE users SET assets = assets + ? WHERE id = 1 AND role = 1"
					_, err = tx.Exec(query, con.Total*0.002)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "收取手续费失败"})
						return
					}
					goto Loop
				}
			}
		}
	case 2:
		// 冻结可用资产
		query = "UPDATE user_treasure SET remain = remain - ? WHERE user_id = ? AND treasure_id = ? AND remain >= ?"
		_, err = tx.Exec(query, con.Amount, userID, con.TreasureID, con.Amount)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "冻结可用资产失败"})
			return
		}
		// 查询匹配
		query = "SELECT id, user_id, price, (amount-trade_amount) AS num FROM consign " +
			"WHERE treasure_id = ? AND type = ? AND price >= ? AND amount-trade_amount > ? ORDER BY price desc, create_time asc"
		rows, err := db.Queryx(query, con.TreasureID, 1, con.Price, 0)
		if err == nil {
			type Arr struct {
				Cid   int64   `db:"id"`
				Uid   int64   `db:"user_id"`
				Price float64 `db:"price"`
				Num   float64 `db:"num"`
			}
			var arrs []Arr
			for rows.Next() {
				var arr Arr
				err = rows.StructScan(&arr)
				if err != nil {
					c.JSON(200, gin.H{"code": 401, "msg": "获取数据失败"})
					return
				}
				arrs = append(arrs, arr)
			}
			// 计算价格浮动
			for _, v := range arrs {
				if price > 0 {
					order.PriceFloat = (v.Price - price) / price
				} else {
					order.PriceFloat = 0
				}
				if con.Amount < v.Num { // 卖出数量小于买入数量
					// 买入
					query = "UPDATE consign SET trade_amount = trade_amount + ?, trade_time = ? WHERE id = ?"
					_, err = tx.Exec(query, con.Amount, con.CreateTime, v.Cid)
					if err != nil {
						c.Error(err).SetType(gin.ErrorTypePrivate)
						//c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+1)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成买入订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, con.Amount, v.Uid, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, mcid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+2)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成卖出订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, v.Price*con.Amount, userID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": " 增加卖方资金失败"})
						return
					}
				} else if con.Amount == v.Num { // 卖出数量等于买入数量
					// 买入
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, v.Cid)
					if err != nil {
						c.Error(err).SetType(gin.ErrorTypePrivate)
						//c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+3)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成买入订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, con.Amount, v.Uid, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ?  WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, mcid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+4)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.Error(err).SetType(gin.ErrorTypePrivate)
						c.JSON(200, gin.H{"code": 401, "msg": "生成卖出订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, v.Price*con.Amount, userID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": " 增加卖方资金失败"})
						return
					}
				} else { // 卖出数量大于买入数量
					// 买入
					query = "UPDATE consign SET trade_amount = amount, trade_time = ?, status = ? WHERE id = ?"
					_, err = tx.Exec(query, con.CreateTime, 2, v.Cid)
					if err != nil {
						c.Error(err).SetType(gin.ErrorTypePrivate)
						//c.JSON(200, gin.H{"code": 401, "msg": "买入失败"})
						return
					}
					// 生成买入订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+5)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, v.Uid, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						1, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, userID, time.Now().Format("2006-01-02"))
					if err != nil {
						c.Error(err).SetType(gin.ErrorTypePrivate)
						c.JSON(200, gin.H{"code": 401, "msg": "生成卖出订单失败"})
						return
					}
					// 添加买方资产
					query = "UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?"
					_, err = tx.Exec(query, v.Num, v.Uid, con.TreasureID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "添加资产失败"})
						return
					}
					// 卖出
					query = "UPDATE consign SET trade_amount = trade_amount + ?, trade_time = ? WHERE id = ?"
					_, err = tx.Exec(query, v.Num, con.CreateTime, mcid)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "卖出失败"})
						return
					}
					// 生成卖出订单
					order.OrderNumber = strconv.FormatInt(time.Now().Unix(), 10) + fmt.Sprintf("%03v", rnd.Int31n(1000)+6)
					query = "INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
					_, err = tx.Exec(query, nil, userID, con.TreasureID, time.Now().Format("15:04:05"), order.OrderNumber, order.ImageID,
						2, v.Price, con.Amount, v.Price*con.Amount, 1, order.PriceFloat, v.Uid, time.Now().Format("2006-01-02"))
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "生成卖出订单失败"})
						return
					}
					// 增加卖方资金
					query = "UPDATE users SET assets = assets + ? WHERE id = ?"
					_, err = tx.Exec(query, v.Price*v.Num, userID)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": " 增加卖方资金失败"})
						return
					}
				}
				if con.Amount-v.Num > 0 {
					con.Amount = con.Amount - v.Num
				} else {
					// 扣除手续费
					query = "UPDATE users SET assets = assets - ? WHERE id = ? AND assets >= ?"
					_, err = tx.Exec(query, con.Total*0.002, userID, con.Total*0.002)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "扣除手续费失败"})
						return
					}
					// 收取手续费
					query = "UPDATE users SET assets = assets + ? WHERE id = 1 AND role = 1"
					_, err = tx.Exec(query, con.Total*0.002)
					if err != nil {
						c.JSON(200, gin.H{"code": 401, "msg": "收取手续费失败"})
						return
					}
					goto Loop
				}
			}
		}
		defer rows.Close()
	}
Loop:
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "提交失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "委托成功"})
}

// 委托列表
func listConsign(c *gin.Context) {
	type Consign struct {
		ID         int64   `db:"id"          json:"id"`
		CreateTime string  `db:"create_time" json:"createTime"`
		Type       int     `db:"type"        json:"type"`
		Price      float64 `db:"price"       json:"price"`
		Amount     float64 `db:"amount"      json:"amount"`
		Total      float64 `db:"total"       json:"total"`
		Status     int     `db:"status"      json:"status"`
	}
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	TreasureID := c.Query("treasureID")
	status := c.Query("status")
	Me, err := strconv.ParseBool(c.Query("me"))
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误me(bool)"})
		return
	}
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "10"), 10, 64)
	db := c.MustGet("DB").(*sqlx.DB)
	var query string
	var args []interface{}
	var total int64
	if !Me {
		typ, _ := strconv.ParseInt(c.Query("type"), 10, 64)
		if status == "1" {
			// 委托中
			query = "SELECT id, create_time, type, price,sum(amount-trade_amount) AS amount, status, " +
				"price*(sum(amount-trade_amount)) AS total FROM consign " +
				"WHERE treasure_id = ? AND type = ? AND status = ? GROUP BY price ORDER BY price desc"
			args = append(args, TreasureID, typ, status)
		} else {
			// 已完成或撤销的
			query = "SELECT id, create_time, type, price, amount, status, total FROM consign " +
				"WHERE treasure_id = ? AND type = ? AND status IN(?) ORDER BY price desc"
			args = append(args, TreasureID, typ, status)
		}

	} else {
		if !ok {
			c.JSON(200, gin.H{"code": 401, "msg": "尚未登录"})
			return
		}
		// 自己的
		if status == "1" {
			query = "SELECT id, create_time, type, price, amount, status, total FROM consign WHERE user_id = ? "
			sql := "SELECT COUNT(id) FROM consign WHERE user_id = ? "
			args = append(args, userID)
			if TreasureID != "" {
				query += "AND treasure_id = ? "
				args = append(args, TreasureID)
				sql += "AND treasure_id = ? "
			}
			sql += "AND status = 1"
			db.QueryRow(sql, args...).Scan(&total)
			query += "AND status = 1 ORDER BY create_time desc LIMIT ?, ?"
			args = append(args, (page-1)*offset, offset)
		} else {
			query = "SELECT id, create_time, type, price, amount, status, total FROM consign WHERE user_id = ? "
			sql := "SELECT COUNT(id) FROM consign WHERE user_id = ?"
			args = append(args, userID)
			if TreasureID != "" {
				query += "AND treasure_id = ? "
				args = append(args, TreasureID)
				sql += "AND treasure_id = ? "
			}
			if status != "" {
				query += "AND status IN(?) "
				args = append(args, status)
				sql += "AND status IN(?) "
			}
			sql += "ORDER BY create_time desc"
			db.QueryRow(sql, args...).Scan(&total)
			query += "ORDER BY create_time desc LIMIT ?, ?"
			args = append(args, (page-1)*offset, offset)
		}
	}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "暂无数据"})
		return
	}
	defer rows.Close()
	var lists []Consign
	for rows.Next() {
		var con Consign
		err = rows.StructScan(&con)
		if err != nil {
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
		lists = append(lists, con)
	}
	if TreasureID != "" {
		type Data struct {
			Price      float64 `json:"price"`
			Assets     float64 `json:"assets"`
			BuyNum     float64 `json:"buyNum"`
			SellNum    float64 `json:"sellNum"`
			Frequency  int64   `json:"frequency"`
			SquareAll  float64 `json:"squareAll"`
			Transfer   int64   `json:"transfer"`
			Rate       float32 `json:"rate"`
			PriceFloat float32 `json:"priceFloat"`
		}
		var data Data
		// 最新成交价格
		query = "SELECT price, price_float FROM orders WHERE treasure_id = ? ORDER BY create_time desc, id desc LIMIT 1"
		db.QueryRow(query, TreasureID).Scan(&data.Price, &data.PriceFloat)
		// 账户可用余额
		query = "SELECT assets FROM users WHERE id = ?"
		db.QueryRow(query, userID).Scan(&data.Assets)
		// 可卖出量
		query = "SELECT remain FROM user_treasure WHERE user_id = ? AND treasure_id = ?"
		db.QueryRow(query, userID, TreasureID).Scan(&data.SellNum)
		// 可买入量
		query = "SELECT h.acreage FROM treasure AS t INNER JOIN house AS h ON h.id = t.house_id WHERE t.id = ?"
		db.QueryRow(query, TreasureID).Scan(&data.BuyNum)
		data.BuyNum = data.BuyNum - data.SellNum
		// 交易次数
		query = "SELECT COUNT(id), SUM(amount) FROM consign WHERE treasure_id = ? AND status = ?"
		db.QueryRow(query, TreasureID, 2).Scan(&data.Frequency, &data.SquareAll)
		c.JSON(200, gin.H{"code": 200, "data": lists, "datas": data, "total": total})
	} else {
		c.JSON(200, gin.H{"code": 200, "data": lists, "total": total})
	}

}

// 撤单
func putConsign(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误id(int)"})
		return
	}
	typ, err := strconv.ParseInt(c.Query("type"), 10, 0)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误type(int)"})
		return
	}
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "尚未登录"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	tx, err := db.Begin()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "事物开启失败)"})
		return
	}
	defer tx.Rollback()
	var tid int64
	var amount, trade_amount, price float64
	query := "SELECT amount, trade_amount, price, treasure_id FROM consign WHERE id = ? AND status = ?"
	err = tx.QueryRow(query, cid, 1).Scan(&amount, &trade_amount, &price, &tid)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "无法撤单，请联系管理员"})
		return
	}
	switch typ {
	case 1:
		// 买方撤单
		_, err = tx.Exec("UPDATE users SET assets = assets + ? WHERE id = ?", (amount-trade_amount)*price, userID)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "撤回资金失败"})
			return
		}
	case 2:
		// 卖方撤回
		_, err = tx.Exec("UPDATE user_treasure SET remain = remain + ? WHERE user_id = ? AND treasure_id = ?", amount-trade_amount, userID, tid)
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "撤回资产失败"})
			return
		}
	}
	_, err = tx.Exec("UPDATE consign SET status = -1 WHERE id = ?", cid)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "撤回委托失败"})
		return
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "撤单失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "撤单成功,未交易部分已退回原账户，请查收"})
}

// 删除委托
func deleteConsign(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("uid").(int64)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "请先登录"})
		return
	}
	cid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数错误"})
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	query := "DELETE FROM consign WHERE id = ? AND user_id = ? AND status <> ?"
	_, err = db.Exec(query, cid, userID, 1)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "删除成功"})
}

// K线图
func KLine(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	tid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"code": 401, "msg": "参数有误"})
		return
	}
	var data []interface{}
	// 查询时间
	query := "SELECT create_date, price FROM orders WHERE treasure_id = ? GROUP BY create_date ORDER BY create_date asc"
	rows, _ := db.Queryx(query, tid)
	defer rows.Close()
	for rows.Next() {
		var date, price string
		rows.Scan(&date, &price)
		// data:开盘时间 开盘价 封盘价 最高价 最低价
		// 初始时间
		var mint string
		query = "SELECT MIN(create_time) FROM orders WHERE treasure_id = ? AND create_date = ?"
		db.QueryRow(query, tid, date).Scan(&mint)
		// 结束时间
		var maxt string
		query = "SELECT MAX(create_time) FROM orders WHERE treasure_id = ? AND create_date = ?"
		db.QueryRow(query, tid, date).Scan(&maxt)
		// 开盘价
		var sprice string
		query = "SELECT price FROM orders WHERE treasure_id = ? AND create_date = ? AND create_time = ? GROUP BY create_time"
		db.QueryRow(query, tid, date, mint).Scan(&sprice)
		// 封盘价
		var tprice string
		query = "SELECT price FROM orders WHERE treasure_id = ? AND create_date = ? AND create_time = ? GROUP BY create_time"
		db.QueryRow(query, tid, date, maxt).Scan(&tprice)
		// 最高和最低价
		query = "SELECT MAX(price), MIN(price) FROM orders WHERE treasure_id = ? AND create_date = ?"
		var hprice, lprice string
		db.QueryRow(query, tid, date).Scan(&hprice, &lprice)
		data = append(data, []string{date, sprice, tprice, lprice, hprice})
	}
	if data == nil {
		var price string
		date := time.Now().Format("2006-01-02")
		query = "SELECT h.evaluate FROM treasure AS t INNER JOIN house AS h ON t.house_id = h.id WHERE t.id = ?"
		db.QueryRow(query, tid).Scan(&price)
		data = append(data, []string{date, price, price, price, price, price})
	}
	c.JSON(200, gin.H{"code": 200, "data": data})
}
