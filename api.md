## API ##

### 用户注册 ###

POST /api/user

``` json
{
	"username":"username", 用户名
	"password":"password", 密码
	"mailbox":"mailbox", 邮箱
	"fundsPassword":"fundsPassword", 资金密码
	"inviteCode":"612922", 邀请码
	"code":"612922" 验证码
}
```

返回

``` json
{
  "code": 200,
  "data": "注册成功"
}
```

### 用户登录 ###

POST /api/session

``` json
{
	"username":"username", 用户名
	"password":"password", 密码
	"code":"630658", 验证码
	"bindIp":true/false, 是否绑定IP
	"status": true/false true管理员登录false普通用户
}
```

返回

``` json
{
  "code": 200,
  "data": {
        "id": 1,
        "username": "admin", 用户名
        "grade": 2, vip等级
        "phone":"15646878"
    }
}
```

### 注销登录 ###

GET /api/logout

返回

``` json
{
    "code": 200,
    "data": "注销成功"
}
```

### 修改密码 ###

PUT /api/user
{
    "username":"123", 账号 可传（用户名/邮箱/手机号）
    "oldPassword":"456654", 原密码 可传
    "newPassword":"456546", 新密码 必传
    "code":"465654", 验证码 必传
    "type":true 类型 可传 true 修改资金密码 false 修改登录密码 默认(false)
}
返回

``` json
{
    "code": 200,
    "data": "密码修改成功"
}
```

### 验证码获取 ###

GET /api/verify

返回

``` json
{
    "code": 200,
    "data": "324945"
}
```

### 验证码发送 ###

GET /api/sendVerify?name=mail&mail=abc@qq.com
{
    name 必选  例mail/sms   mail邮箱/sms短信
    mail 可选  例123@qq.com  name为mail 时作为接收者
    phone 可选 例13800280500 name为sms 时作为接收者
    type 可选 注：如果name为sms 必选 模板类型1.信息变更2.身份验证3.登录确认4.修改密码
}
返回
``` json
{
    "code": 200,
    "data": "发送成功！"
}
```

### 图片上传 ###

POST /api/image

``` json
{
	"file":"file"
}
```

返回

``` json
{
    "code": 200,
    "data": 1
}
```

### 图片上传 ###

GET /api/image/:id

### 图片删除 ###

DELETE /api/image/:id

### 身份验证 ###

POST /api/validate

``` json
{
	"realName":"admin", 真实姓名
	"idCard":"654321", 身份证号
	"proct":"proct", 省市区
	"address":"address", 详细地址
	"sex":1, 性别1男2女
	"birthday":"birthday", 出生日期
	"qq":"qq", QQ
	"phone":"phone", 手机号
	"spare":"spare", 备用号码
	"handAccount":1, 手持账号照
	"type":1, 身份证类型1二代身份2临时
	"frontimg":1, 身份证正面
	"oppositeimg":1, 身份证背面
	"handCard":1 手持身份证
}
```

返回

``` json
{
    "code": 200,
	"data": "提交成功"
}
```

### 身份验证列表 ###

GET /api/validate 可选参数[page string,[type string]]

返回

``` json
{
	"code":200,
	"data":{
		{
			"id":"id",
			"realName":"realName", 真实姓名
			"idCard":"idCard", 身份证号
			"handAccount":1, 手持账号照
			"type":1, 身份证类型1二代身份2临时
			"frontimg":1, 身份证正面
			"oppositeimg":1, 身份证背面
			"handCard":1, 手持身份证
			"status":1 状态:0审核中1通过2失败
		}
	}

}
```
### 查看身份验证 ###

GET /api/show_validate?id=1 可选参数id,默认查己

返回

``` json
{
	"code":200,
	"data":{
		"id":"id",
		"realName":"realName", 真实姓名
		"idCard":"idCard", 身份证号
		"proct":"proct", 省市区
		"address":"address", 详细地址
		"sex":1, 性别1男2女
		"birthday":"birthday", 出生日期
		"qq":"qq", QQ
	    "phone":"phone", 手机号
	    "spare":"spare", 备用号码
		"handAccount":1, 手持账号照
		"type":1, 身份证类型1二代身份2临时
		"frontimg":1, 身份证正面
		"oppositeimg":1, 身份证背面
		"handCard":1, 手持身份证
		"status":1 状态:0审核中1通过2失败
	}

}
```

### 审核身份认证 ###

PUT /api/validate

``` json
{
	"id":1,
	"status":1  状态0未审核1审核通过2未通过
}
```

返回

``` json
{
	"code":200,
	"data":"操作成功"

}
```

### 删除身份认证 ###

DELETE /api/validate/:id

返回

``` json
{
	"code":200,
	"data":"删除成功"

}
```

### 上传宝贝 ###

POST /api/treasure

``` json
{
	"sellway":1, 出让方式1.全资出售2.面积交易3.借贷
	"openbusiness":"654321", 开盘商
	"opentime":"12:00", 开盘时间
	"housecard":"1", 房产证号
	"housetype":1, 房产质押方式1.已经质押
	"propertyright":1, 产权
	"oneprice":11.50, 一口价
	"decorationfree":11.50, 装修费
	"title":"qw", 标题
	"describe":"dsa", 描述
	"contacts":"ads", 联系人
	"phone":"dasd", 联系电话
	"type":1, 宝贝类别1.合租房2.二手房3.商铺租售4.厂房5.鞋子楼6.仓库7.土地8.一口价
	"contract":"contract", 合约
	"houseDate":{ 房子信息
		"roomnum":1, 室
		"officenum":1, 厅
		"toiletnum":1, 卫
		"orientation":1, 朝向0.不限1.东2.南3.西4.北5.南北6.东西7.东南8.西南9.东北10.西北
		"decoration":1, 装修风格0.不限1.毛坯房2.简单装修3.中等装修4.精装修5.豪华装修
		"type":1, 房屋类型0不限1商品房
		"acreage":12.3, 房屋面积
		"height":14.5, 楼高
		"floor":1, 楼层
		"adress":"ds ad sa asd", 房屋地址
		"evaluate":11.50, 市场估价
		"images":[1,2,3] 房子图片
	}
}
```

返回

``` json
{
    "code": 200,
    "data": "上传成功"
}
```

### 宝贝列表 ###

GET /api/treasure

``` string

?page=1&offset=10&owner=true&province="浙江省"&city="温州市"&area="乐清市"
可选参数
	page 当前页 默认第1页
	offset 每页几条 默认10条
	owner true个人/false全部 默认false
	province 省份
	city 城市
	area 区域
	status 宝贝状态1上架2下架3审核中
	search 查询条件
```
返回

``` json
{
    "code": 200,
    "data": [
        {
            "id": 3, 宝贝ID
            "evaluate": 11.5, 市场估价
            "priceFloat": 0, 价格浮动
            "title": "qczxcxzw", 标题
            "acreage": 12.3, 房屋面积
            "remain": 12.3, 剩余面积
            "adress": "sda", 房屋地址
            "imageId": 7, 图片ID
			"status":1, 出售状态1在售2关闭
			"priceFloat":0.01% 价格浮动
        }
    ],
    "total": 1 总条数
}
```

### 宝贝详情 ###

GET /api/treasure/:id

返回

``` json
{
    "code": 200,
    "data": {
        "id": 1, //宝贝ID
        "houseId": 1, //房子ID
        "sellway": 1, //出让方式1.全资出售2.面积交易3.借贷
        "openbusiness": "654321", //开盘商
        "opentime": "12:00", //开盘时间
        "housecard": "1", //房产证号
        "housetype": 1, //房产质押方式1.已经质押
        "propertyright": 1,  //产权
        "evaluate": 11.5, //市场估价
        "oneprice": 11.5, //一口价
        "decorationfree": 11.5, //装修费
        "title": "qw", //标题
        "describe": "dsa", //描述
        "contacts": "ads", //联系人
        "phone": "dasd", //联系电话
		"type":1, //宝贝类别1.合租房2.二手房3.商铺租售4.厂房5.鞋子楼6.仓库7.土地8.一口价
        "remain": 80, //剩余面积
        "priceFloat": 0, //价格浮动
        "status":1, //是否在售1是2否
        "contract":"contract", 合约
        "houseDate": {//房子信息
            "roomnum": 1, //室
            "officenum": 1, //厅
            "toiletnum": 1, //卫
            "orientation": 1, //朝向0.不限1.东2.南3.西4.北5.南北6.东西7.东南8.西南9.东北10.西北
            "decoration": 1, //装修风格0.不限1.毛坯房2.简单装修3.中等装修4.精装修5.豪华装修
            "type": 1, //房屋类型0不限1商品房
            "acreage": 12.3, //房屋面积
            "height": 14.5, //楼高
            "floor": 1, //楼层
            "adress": "dsadsaasd", //房屋地址
			"evaluate": 12.5, //市场估价
            "images": [ //房子图片
                1,
                2,
                3
            ]
        }
    }
}
```

### 审核宝贝 ###

PUT /api/treasure

``` json
{
	"id":1,
	"status":1  状态1通过2不通过3审核中
}
```

### 宝贝删除 ###

DELETE /api/treasure?id=? 支持批量删除(id=1,2,3...)

返回
``` json
{
    "code": 200,
    "data": "删除成功"
}
```

### 修改宝贝 ###

PUT /api/treasure/:id   ID 随便写

``` json
{
    "id":1, 宝贝ID
	"sellway":1, 出让方式1.全资出售2.面积交易3.借贷
	"openbusiness":"654321", 开盘商
	"opentime":"12:00", 开盘时间
	"housecard":"1", 房产证号
	"housetype":1, 房产质押方式1.已经质押
	"propertyright":1, 产权
	"oneprice":11.50, 一口价
	"decorationfree":11.50, 装修费
	"title":"qw", 标题
	"describe":"dsa", 描述
	"contacts":"ads", 联系人
	"phone":"dasd", 联系电话
	"type":1, 宝贝类别1.合租房2.二手房3.商铺租售4.厂房5.鞋子楼6.仓库7.土地8.一口价
	"contract":"contract", 合约
	"houseDate":{ 房子信息
		"roomnum":1, 室
		"officenum":1, 厅
		"toiletnum":1, 卫
		"orientation":1, 朝向0.不限1.东2.南3.西4.北5.南北6.东西7.东南8.西南9.东北10.西北
		"decoration":1, 装修风格0.不限1.毛坯房2.简单装修3.中等装修4.精装修5.豪华装修
		"type":1, 房屋类型0不限1商品房
		"acreage":12.3, 房屋面积
		"height":14.5, 楼高
		"floor":1, 楼层
		"adress":"ds ad sa asd", 房屋地址
		"evaluate":11.50, 市场估价
		"houseimg":"1,2,3,4",  图片ID
		"images":[1,2,3] 房子图片
	}
}
```

### 创建委托 ###

POST /api/consign

``` json
{
	"treasureID":1, //宝贝ID
	"type":1, //类型 1买入 2卖出
	"price":12.5, //单价
	"amount":80, //数量
	"total":1000 //总价
}
```

返回
``` json
{
    "code": 200,
    "data": "委托成功"
}
```

### 订单列表 ###

POST /api/order

```可选参数
{
	page:1 当前页
	type:1 类型1买入2卖出
	lenth:10 每页条数
	title:"商品标题"
	orderNumber:"订单编号"
	tradeTime:"开始时间,结束时间"
	evaluate:0 评价状态(暂未开放 默认0)
	status:1 交易状态 1交易成功2交易中3交易关闭-1.已删除
}
```

返回
``` json
{
  	"id":1, 订单ID
	"treasureID":1, 宝贝ID
	"createDate":1, 创建时间
	"orderNumber":1, 订单号
	"imageID":1, 图片ID
	"type":1, 类型1买入2卖出
	"price":1, 价格
	"amount":1, 数量
	"realPay":1, 实付款
	"status":1, 订单状态 1.交易成功2.交易中3.交易关闭
	"title":1, 标题
	"roomnum":1, 室
	"officenum":1, 厅
	"toiletnum":1, 卫
	"orientation":1, 朝向0.不限1.东2.南3.西4.北5.南北6.东西7.东南8.西南9.东北10.西北
	"decoration":1, 装修风格0.不限1.毛坯房2.简单装修3.中等装修4.精装修5.豪华装修
	"houseType":1, 房屋类型0不限1商品房
	"acreage":1, 房屋面积
	"height":1, 楼高
	"floor":1, 楼高
	"adress":1 房屋地址
}
```

### 删除订单 ###

delete /api/order?id=? 支持批量删除(id=1,2,3...)

返回
``` json
{
    "code": 200,
    "data": "删除成功"
}
```

### 委托列表 ###

GET /api/consign

``` param
?treasureID=1&type=1&me=true&status=1
{
	treasureID 	//宝贝ID
	type		// 1买入2卖出
	status  //1交易中2已完成-1撤单
	me			//true自己false全部
}
```

返回
``` json
{
    "code": 200,
    "data": [
        {
            "id": 1,
            "createTime": "2018-09-07 13:21:24", 挂单时间
            "type": 1, 类型1买入2卖出
            "price": 13, 价格
            "amount": 12, 数量
            "total": 156, 总价
            "status": 1.委托中 2.已完成 -1.撤销
        }
    ],
    "datas": [
		{
			"price":1, 最新价格
			"assets":2, 账户余额
			"buyNum":3, 可买入量
			"sellNum":4, 可卖出量
			"frequency":0, 交易次数
			"squareAll":0, 交易累计平方
			"transfer":0, 过户次数
			"rate":0, 过户占比
			"priceFloat":0.01 价格浮动
		}
	],
	"total":10  总条数
}
```

### 撤回委托 ###

PUT /api/consign?id=1&type=1

id 挂单ID type 买入1 卖出2

返回
``` json
{
    "code": 200,
    "data": "撤单成功,未交易部分已退回原账户，请查收"
}
```

### 删除委托 ###

DELETE /api/consign/:id

id 挂单ID

返回
``` json
{
    "code": 200,
    "data": "删除成功"
}
```

### K线 ###

GET /api/KLine/:id

id 宝贝ID

返回
``` json
{
    "code": 200,
    "data": [
        [
            "2018-09-14", 日期
            "200.000000", 开盘价
            "200.000000", 收盘价
            "200.000000", 最高价
            "200.000000"  最低价
        ],
        [
            "2018-09-15", 日期
            "200.000000", 开盘价
            "200.000000", 收盘价
            "200.000000", 最高价
            "200.000000"  最低价
        ]
    ]
}
```

### 删除资产 ###

DELETE /api/assets/:id

id 宝贝ID

返回
``` json
{
    "code": 200,
    "data": "删除成功"
}
```

### 资产列表 ###

GET /api/assets

返回
``` json
{
    "code": 200,
    "square":square,  总平方
	"RMB":RMB, 总人民币
	"squareCoin":squareCoin, 总平方币
	"assets": 5000, 我的账户资金
    "data": [
        {
            "id": 3,  // 排序ID
            "treasureID": 1, // 宝贝ID
            "remian": 50, //资产数量
            "address": "浙江省温州市乐清市浙江省温州市乐清市", // 资金币种
            "price": 100, // 最新价格
            "total": 5000 // 估算总价
        }
    ]
}
```

### 更新手机号 ###

PUT /api/modPhone
``` json
{
    "phone":"13800280500",
    "code":"456123"
}
返回
``` json
{
    "code": 200,
    "data": "更新成功!",
    "phone":"13800280500"
}
```
### 设置短信接口 ###

POST /api/setDysms
``` json
{
    "id":1,   模板ID （可选） 带ID表示更新 不带ID表示新增
    "keyid":"LTAIZGnJihiQxN6t", accessKeyID
    "secret":"40KYuFNn1VONELfvsaN5WA7PTEuRjI", accessSecret
    "signname":"全民豆豆", 短信签名
    "code":"SMS_145885447", 短信模板Code
    "status":1, 是否开启 1是2否
    "type":1  模板类型1.信息变更2.身份验证3.登录注册
}
返回
``` json
{
    "code": 200,
    "data": "更新成功!"   更新成功!/新增成功!
}
```
### 设置邮箱接口 ###

POST /api/setMail
``` json
{
    "id":1,   模板ID （可选） 带ID表示更新 不带ID表示新增
    "username":"236491572@qq.com", 用户名
    "password":"vktraeoawyilbjac", 密码
    "nickname":"平方链", 昵称
    "subject":"获取邮箱验证码 - 平方链", 邮件标题
    "content":"验证码为：%s。请尽快填写。如果不是您操作的本次邮件，请忽略。", 发送内容
    "status":1  状态1开启2关闭
}
返回
``` json
{
    "code": 200,
    "data": "更新成功!"   更新成功!/新增成功!
}
```
### 短信模板列表 ###

POST /api/dysmsList

返回
``` json
{
    "code": 200,
    "data": [
        "id":1,   模板ID （可选） 带ID表示更新 不带ID表示新增
        "keyid":"LTAIZGnJihiQxN6t", accessKeyID
        "secret":"40KYuFNn1VONELfvsaN5WA7PTEuRjI", accessSecret
        "signname":"全民豆豆", 短信签名
        "code":"SMS_145885447", 短信模板Code
        "status":1, 是否开启 1是2否
        "type":1  模板类型1.信息变更2.身份验证3.登录注册
    ]
}
```
## 邮箱模板列表 ###

POST /api/mailList

返回
``` json
{
    "code": 200,
    "data": [
            "id":1,   模板ID （可选） 带ID表示更新 不带ID表示新增
            "username":"236491572@qq.com", 用户名
            "password":"vktraeoawyilbjac", 密码
            "nickname":"平方链", 昵称
            "subject":"获取邮箱验证码 - 平方链", 邮件标题
            "content":"验证码为：%s。请尽快填写。如果不是您操作的本次邮件，请忽略。", 发送内容
            "status":1  状态1开启2关闭
    ]
}
```
## 删除短信/邮件模板 ###

DELETE /api/delTemp?name=mail&id=1
``` param
    name   mail/sms  区别类型：mail邮箱/sms短信
    id  1   邮箱ID/短信ID

返回
``` json
{
    "code": 200,
    "data": "删除成功!"
}
```