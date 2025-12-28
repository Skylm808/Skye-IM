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