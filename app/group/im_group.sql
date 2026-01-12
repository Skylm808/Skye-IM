-- 群组表
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
