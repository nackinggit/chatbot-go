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

CREATE TABLE `chat_history_event` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `userid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'user id',
    `botid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'bot id',
    `start_msgid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '记录开始id',
    `end_msgid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '记录开始id',
    `date_str` char(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '事件时间',
    `event` text COLLATE utf8mb4_unicode_ci COMMENT '事件描述',
    `addr` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ' ' COMMENT '地点',
    `todo` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '  ' COMMENT '代办项',
    `emotional` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ' ' COMMENT '情绪状态',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '聊天记录事件表';

INSERT INTO
    llm_model(
        `bind_im_bot_id`,
        `bind_im_bot_name`,
        `api`,
        `model_key`,
        `model_name`
    )
VALUES
    (
        29,
        '鬼灭-灶门炭治郎',
        'doubao',
        'bot-20241022151352-jqtlz',
        'doubao-1.5'
    ),
    (
        7,
        '抽象大师',
        'doubao',
        'bot-20241006224855-zlvxs',
        'doubao-1.5'
    ),
    (
        30,
        '鬼灭之刃--剧情三千问',
        'coze',
        '7433401037107036169',
        'doubao-1.5'
    ),
    (
        36,
        '摆烂大师',
        'doubao',
        'bot-20241006160324-plxnr',
        'doubao-1.5'
    ),
    (
        37,
        '鬼灭--嘴平伊之助',
        'doubao',
        'bot-20241127220024-8s69p',
        'doubao-1.5'
    ),
    (
        38,
        '鬼灭-我妻善逸',
        'doubao',
        'bot-20241127215558-vldkg',
        'doubao-1.5'
    ),
    (
        39,
        '鬼灭-灶门祢豆子',
        'doubao',
        'bot-20241127214805-fcbpv',
        'doubao-1.5'
    ),
    (
        45,
        '小宝-多模态',
        'coze',
        '7453660789518909475',
        'doubao-1.5'
    );