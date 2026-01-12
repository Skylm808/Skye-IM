-- 群成员表
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
