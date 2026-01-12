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
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_msg_id` (`msg_id`),
    KEY `idx_from_user` (`from_user_id`),
    KEY `idx_to_user` (`to_user_id`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_conversation` (`from_user_id`, `to_user_id`, `created_at`),
    KEY `idx_unread` (`to_user_id`, `status`, `created_at`),
    KEY `idx_group_seq` (`group_id`, `seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM 消息主表(支持私聊与群聊)';
ALTER TABLE `im_message` 
ADD COLUMN `at_user_ids` TEXT COMMENT '被@的用户ID列表,JSON格式,如["123","456"],@all用特殊值"-1"';
-- 添加索引以优化查询@我的消息
ALTER TABLE `im_message`
ADD INDEX `idx_at_users` (`group_id`, `chat_type`) USING BTREE;