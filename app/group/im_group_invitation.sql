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