-- 消息表：存储私聊消息
CREATE TABLE IF NOT EXISTS `im_message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
    `msg_id` VARCHAR(64) NOT NULL COMMENT '消息唯一标识(UUID)',
    `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    `to_user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
    `content` TEXT NOT NULL COMMENT '消息内容',
    `content_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型: 1-文字 2-图片 3-文件 4-语音',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '消息状态: 0-未读 1-已读 2-撤回',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_msg_id` (`msg_id`),
    KEY `idx_from_user` (`from_user_id`),
    KEY `idx_to_user` (`to_user_id`),
    KEY `idx_conversation` (`from_user_id`, `to_user_id`, `created_at`),
    KEY `idx_unread` (`to_user_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='私聊消息表';
