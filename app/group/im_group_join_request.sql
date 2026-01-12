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