-- IM系统用户表
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `im_auth` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `im_auth`;

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

