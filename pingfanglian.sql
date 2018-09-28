/*
 Navicat Premium Data Transfer

 Source Server         : pingfanglian
 Source Server Type    : MySQL
 Source Server Version : 100126
 Source Host           : dev.lq:3306
 Source Schema         : pingfanglian

 Target Server Type    : MySQL
 Target Server Version : 100126
 File Encoding         : 65001

 Date: 27/09/2018 17:21:29
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for consign
-- ----------------------------
DROP TABLE IF EXISTS `consign`;
CREATE TABLE `consign`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `treasure_id` int(11) NOT NULL COMMENT '宝贝ID',
  `create_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '挂单时间',
  `type` tinyint(2) NOT NULL COMMENT '类型1买入2卖出',
  `price` decimal(12, 6) NOT NULL COMMENT '价格',
  `amount` decimal(12, 2) NOT NULL COMMENT '挂单量',
  `status` tinyint(2) NOT NULL DEFAULT 1 COMMENT '1.委托中 2.已完成 -1.撤销',
  `total` decimal(12, 6) NOT NULL COMMENT '总价',
  `trade_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '成交时间',
  `trade_amount` decimal(12, 2) NOT NULL COMMENT '成交量',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 98 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '宝贝' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for dysms
-- ----------------------------
DROP TABLE IF EXISTS `dysms`;
CREATE TABLE `dysms`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `keyid` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'accessKeyID',
  `secret` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'accessSecret',
  `signname` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '短信签名',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '短信模板Code',
  `status` tinyint(2) NOT NULL DEFAULT 2 COMMENT '开启状态1开启2关闭',
  `type` tinyint(4) NOT NULL COMMENT '模板类型1.信息变更2.身份验证3.登录确认4.修改密码',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for house
-- ----------------------------
DROP TABLE IF EXISTS `house`;
CREATE TABLE `house`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `roomnum` tinyint(4) NOT NULL COMMENT '室',
  `officenum` tinyint(4) NOT NULL COMMENT '厅',
  `toiletnum` tinyint(4) NOT NULL COMMENT '卫',
  `orientation` tinyint(4) NOT NULL DEFAULT 0 COMMENT '朝向0.不限1.东2.南3.西4.北5.南北6.东西7.东南8.西南9.东北10.西北',
  `decoration` tinyint(4) NOT NULL DEFAULT 0 COMMENT '装修风格0.不限1.毛坯房2.简单装修3.中等装修4.精装修5.豪华装修',
  `type` tinyint(4) NOT NULL COMMENT '房屋类型0不限1商品房',
  `acreage` decimal(10, 2) NOT NULL COMMENT '房屋面积',
  `height` decimal(10, 2) NOT NULL COMMENT '楼高',
  `floor` tinyint(4) NOT NULL COMMENT '楼层',
  `adress` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '房屋地址',
  `evaluate` decimal(10, 6) NOT NULL COMMENT '市场估价',
  `province` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '省份',
  `city` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '城市',
  `area` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '区域',
  `houseimg` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '户型图片',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 14 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '房屋信息' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for house_image
-- ----------------------------
DROP TABLE IF EXISTS `house_image`;
CREATE TABLE `house_image`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `house_id` int(11) NOT NULL COMMENT '房子ID',
  `image_id` int(11) NOT NULL COMMENT '图片ID',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 36 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '房子图片关联表' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for image
-- ----------------------------
DROP TABLE IF EXISTS `image`;
CREATE TABLE `image`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `data` mediumblob NOT NULL COMMENT '图片数据',
  `name` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '图片名称',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 53 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '图片表' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for login_log
-- ----------------------------
DROP TABLE IF EXISTS `login_log`;
CREATE TABLE `login_log`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `last_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '上次登录时间',
  `create_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '当前登录时间',
  `bind_ip` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '登录IP地址',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 84 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '登录日志' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for mail
-- ----------------------------
DROP TABLE IF EXISTS `mail`;
CREATE TABLE `mail`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码',
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '昵称',
  `subject` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '邮件标题',
  `content` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '发送内容',
  `status` tinyint(2) NOT NULL COMMENT '状态1开启2关闭',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for orders
-- ----------------------------
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID 根据类型判断 是卖家还是买家',
  `treasure_id` int(11) NOT NULL COMMENT '宝贝ID',
  `create_time` time(0) NOT NULL COMMENT '下单时间',
  `order_number` varchar(13) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '订单号',
  `image_id` int(11) NOT NULL COMMENT '图片ID',
  `type` tinyint(2) NOT NULL COMMENT '类型：1.买入 2.卖出',
  `price` decimal(12, 6) NOT NULL COMMENT '价格',
  `amount` decimal(12, 2) NOT NULL COMMENT '数量',
  `real_pay` decimal(12, 2) NOT NULL COMMENT '实际支付',
  `status` tinyint(2) NOT NULL DEFAULT 3 COMMENT '订单状态 1.交易成功2.交易中3.交易关闭-1.删除',
  `price_float` decimal(5, 2) NOT NULL DEFAULT 0.00 COMMENT '价格浮动',
  `uid` int(11) NOT NULL COMMENT '用户ID 根据类型判断 是卖家还是买家',
  `create_date` date NOT NULL COMMENT '下单日期',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 32 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '订单表' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for treasure
-- ----------------------------
DROP TABLE IF EXISTS `treasure`;
CREATE TABLE `treasure`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '宝贝上传者',
  `house_id` int(11) NOT NULL COMMENT '房屋ID',
  `sellway` tinyint(4) NOT NULL DEFAULT 1 COMMENT '出让方式1.全资出售2.面积交易3.借贷',
  `openbusiness` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '开盘商',
  `opentime` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '开盘时间',
  `housecard` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '房产证号',
  `housetype` tinyint(4) NOT NULL DEFAULT 1 COMMENT '房产质押方式1.已经质押',
  `propertyright` tinyint(4) NOT NULL COMMENT '产权',
  `oneprice` decimal(10, 2) NOT NULL COMMENT '一口价',
  `decorationfree` decimal(10, 2) NOT NULL COMMENT '装修费',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
  `describe` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '描述',
  `contacts` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '联系人',
  `phone` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '联系电话',
  `status` tinyint(2) NOT NULL DEFAULT 3 COMMENT '状态：1上架2下架3审核中',
  `type` tinyint(4) NOT NULL DEFAULT 1 COMMENT '宝贝类别1.合租房2.二手房3.商铺租售4.厂房5.鞋子楼6.仓库7.土地8.一口价',
  `contract` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '合约',
  `subtitle` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '副标题',
  `create_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '上传时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '宝贝信息' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for user_treasure
-- ----------------------------
DROP TABLE IF EXISTS `user_treasure`;
CREATE TABLE `user_treasure`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `treasure_id` int(11) NOT NULL COMMENT '宝贝ID',
  `remain` decimal(12, 2) NOT NULL COMMENT '宝贝拥有数量',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 28 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户宝贝买卖' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '用户名',
  `password` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '登录密码',
  `mailbox` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '邮箱',
  `funds_password` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '资金密码',
  `invite_code` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '邀请码',
  `role` tinyint(4) NOT NULL DEFAULT 2 COMMENT '身份 1.管理员 2.普通用户 3.vip用户',
  `create_time` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '创建时间',
  `real_name` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '真实姓名',
  `id_card` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '身份证号',
  `assets` decimal(12, 6) NOT NULL DEFAULT 5000.000000 COMMENT '我的资金',
  `grade` tinyint(2) NOT NULL DEFAULT 0 COMMENT 'vip等级',
  `proct` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '省市区',
  `address` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '详细地址',
  `sex` tinyint(1) NOT NULL COMMENT '性别1.男2.女',
  `birthday` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '生日',
  `qq` varchar(24) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT 'QQ',
  `phone` varchar(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '手机号',
  `spare` varchar(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '备用号码',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '平方链用户表' ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for validate
-- ----------------------------
DROP TABLE IF EXISTS `validate`;
CREATE TABLE `validate`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `hand_account` int(11) NOT NULL COMMENT '手持账号照ID',
  `type` tinyint(2) NOT NULL COMMENT '身份证类型 1.二代身份证 2.临时身份证',
  `frontimg` int(11) NOT NULL COMMENT '身份证正面照ID',
  `oppositeimg` int(11) NOT NULL COMMENT '身份证背面照ID',
  `hand_card` int(11) NOT NULL COMMENT '手持身份证照ID',
  `status` tinyint(2) NOT NULL DEFAULT 3 COMMENT '审核状态 3.等待审核 1.审核通过 2.审核不通过',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '身份认证表' ROW_FORMAT = Compact;

SET FOREIGN_KEY_CHECKS = 1;
