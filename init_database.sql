-- ============================================
-- SkyeIM 数据库初始化脚本
-- 版本: 1.0
-- 说明: 包含所有数据表的创建语句
-- 创建时间: 2026-01-18
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `im_auth` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `im_auth`;

-- ============================================
-- 1. 用户认证模块
-- ============================================

-- 用户表
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `username` varchar(64) NOT NULL COMMENT '用户名（唯一）',
    `phone` varchar(20) DEFAULT NULL COMMENT '手机号（唯一）',
    `email` varchar(128) DEFAULT NULL COMMENT '邮箱（唯一）',
    `password` varchar(128) NOT NULL COMMENT '密码（bcrypt加密）',
    `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
    `avatar` varchar(512) DEFAULT '' COMMENT '头像URL',
    `signature` varchar(255) DEFAULT '' COMMENT '个性签名',
    `gender` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '性别：0-未知 1-男 2-女',
    `region` varchar(64) DEFAULT '' COMMENT '地区',
    `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT '状态：1-正常 0-禁用',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_phone` (`phone`),
    UNIQUE KEY `uk_email` (`email`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 2. 好友关系模块
-- ============================================

-- 好友关系表
DROP TABLE IF EXISTS `im_friend`;
CREATE TABLE `im_friend` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `friend_id` bigint unsigned NOT NULL COMMENT '好友ID', 
  `remark` varchar(50) DEFAULT '' COMMENT '好友备注',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1-正常 2-拉黑',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`, `friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友关系表';

-- 好友申请表
DROP TABLE IF EXISTS `im_friend_request`;
CREATE TABLE `im_friend_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `from_user_id` bigint unsigned NOT NULL COMMENT '发起人ID',
  `to_user_id` bigint unsigned NOT NULL COMMENT '接收人ID',
  `message` varchar(200) DEFAULT '' COMMENT '验证消息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_to_user` (`to_user_id`, `status`),
  KEY `idx_from_user` (`from_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友申请表';

-- ============================================
-- 3. 群组管理模块
-- ============================================

-- 群组表
DROP TABLE IF EXISTS `im_group`;
CREATE TABLE `im_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组唯一标识',
  `name` varchar(100) NOT NULL COMMENT '群名称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '群头像',
  `owner_id` bigint NOT NULL COMMENT '群主ID',
  `description` varchar(500) DEFAULT NULL COMMENT '群描述',
  `max_members` int DEFAULT 200 COMMENT '最大成员数',
  `member_count` int DEFAULT 0 COMMENT '当前成员数',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1-正常 2-已解散',
  `invite_confirm_mode` tinyint DEFAULT 1 COMMENT '邀请确认模式: 0-直接加入 1-需要确认（默认）',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_id` (`group_id`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组表';

-- 群成员表
DROP TABLE IF EXISTS `im_group_member`;
CREATE TABLE `im_group_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role` tinyint DEFAULT 3 COMMENT '角色: 1-群主 2-管理员 3-普通成员',
  `nickname` varchar(50) DEFAULT NULL COMMENT '群昵称',
  `mute` tinyint DEFAULT 0 COMMENT '是否禁言: 0-否 1-是',
  `read_seq` BIGINT UNSIGNED DEFAULT 0 COMMENT '已读消息Seq',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群成员表';

-- 群邀请表
DROP TABLE IF EXISTS `im_group_invitation`;
CREATE TABLE `im_group_invitation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `inviter_id` bigint unsigned NOT NULL COMMENT '邀请人ID',
  `invitee_id` bigint unsigned NOT NULL COMMENT '被邀请人ID',
  `message` varchar(200) DEFAULT '' COMMENT '邀请消息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝 3-已过期',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_invitee_status` (`invitee_id`, `status`),
  KEY `idx_inviter` (`inviter_id`),
  KEY `idx_group_invitee` (`group_id`, `invitee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群聊邀请表';

-- 入群申请表
DROP TABLE IF EXISTS `im_group_join_request`;
CREATE TABLE `im_group_join_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `user_id` bigint unsigned NOT NULL COMMENT '申请人ID',
  `message` varchar(200) DEFAULT '' COMMENT '申请理由',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝',
  `handler_id` bigint unsigned DEFAULT NULL COMMENT '处理人ID（群主/管理员）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_group_status` (`group_id`, `status`),
  KEY `idx_user` (`user_id`),
  UNIQUE KEY `uk_group_user_pending` (`group_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='入群申请表';

-- ============================================
-- 4. 消息模块
-- ============================================

-- 消息表
DROP TABLE IF EXISTS `im_message`;
CREATE TABLE IF NOT EXISTS `im_message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键ID',
    `msg_id` VARCHAR(64) NOT NULL COMMENT '消息唯一标识(客户端生成或雪花算法生成的UUID)',
    `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    `to_user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '接收者ID(私聊时有效)',
    `chat_type` TINYINT NOT NULL DEFAULT 1 COMMENT '聊天类型: 1-私聊 2-群聊',
    `group_id` VARCHAR(64) DEFAULT NULL COMMENT '群组ID(群聊时使用)',
    `seq` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息序列号(用于群聊消息连续性校验和拉取偏移量)',
    `content` TEXT NOT NULL COMMENT '消息内容',
    `content_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息内容类型: 1-文字 2-图片 3-文件 4-语音 5-视频',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '消息状态: 0-未读/未处理 1-已读 2-撤回 3-删除',
    `at_user_ids` TEXT COMMENT '被@的用户ID列表,JSON格式,如["123","456"],@all用特殊值"-1"',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_msg_id` (`msg_id`),
    KEY `idx_from_user` (`from_user_id`),
    KEY `idx_to_user` (`to_user_id`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_conversation` (`from_user_id`, `to_user_id`, `created_at`),
    KEY `idx_unread` (`to_user_id`, `status`, `created_at`),
    KEY `idx_group_seq` (`group_id`, `seq`),
    KEY `idx_at_users` (`group_id`, `chat_type`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM 消息主表(支持私聊与群聊)';

-- ============================================
-- 初始化完成提示
-- ============================================
SELECT 'SkyeIM 数据库初始化完成！' AS 'Status';
SELECT '已创建以下数据表：' AS 'Info';
SELECT TABLE_NAME, TABLE_COMMENT 
FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = 'im_auth' 
ORDER BY TABLE_NAME;
