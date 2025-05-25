CREATE TABLE `llm_model` (
    `model_key` varchar(512) PRIMARY KEY NOT NULL COMMENT '模型标识',
    `model_name` varchar(1024) NOT NULL DEFAULT '' COMMENT '模型名称',
    `api` varchar(1024) NOT NULL DEFAULT '' COMMENT '模型api名称',
    `bind_im_bot_name` varchar(1024) NOT NULL DEFAULT '' COMMENT '模型绑定的IM bot名称',
    `bind_im_bot_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '模型绑定的IM botId',
    `is_delete` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除 0-否 1-是',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间'
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '模型信息';

CREATE TABLE IF NOT EXISTS `llm_chat_history` (
    `id` bigint PRIMARY KEY NOT NULL COMMENT 'id',
    `mid` varchar(100) NOT NULL DEFAULT '' COMMENT '消息id',
    `sid` varchar(100) NOT NULL DEFAULT '' COMMENT '会话id',
    `im_user_id` varchar(100) NOT NULL DEFAULT 0 COMMENT '用户id',
    `im_bot_id` varchar(100) NOT NULL DEFAULT 0 COMMENT 'IM机器人id',
    `role` varchar(20) NOT NULL DEFAULT '',
    `message` TEXT COMMENT '消息内容',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    INDEX (`sid`) USING BTREE,
    INDEX (`mid`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '模型信息';