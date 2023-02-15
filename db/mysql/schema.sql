-- db_pridns.forward definition
CREATE TABLE `forward`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `host`        varchar(15)  NOT NULL DEFAULT '' COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `name`        varchar(255) NOT NULL COMMENT '需要转发解析的域名',
    `dns_svr`     varchar(255)          DEFAULT NULL COMMENT '转发目标DNS服务器，可以是多个，多个以逗号分割',
    `deny_global` char(1)      NOT NULL DEFAULT 'N' COMMENT '是否拒绝全局解析',
    `status`      varchar(12)  NOT NULL COMMENT '状态。<br />ENABLE-启用',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_f_host` (`host`),
    KEY `ix_f_name` (`name`)
) COMMENT ='转发配置.';

-- db_pridns.history definition
CREATE TABLE `history`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `name`        varchar(255) NOT NULL COMMENT '需要转发解析的域名',
    `history`     varchar(6000)         DEFAULT NULL COMMENT '解析记录，用于导出使用，多个以逗号分割',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_f_name` (`name`)
) COMMENT ='解析历史.';

-- db_pridns.`domain` definition
CREATE TABLE `domain`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `host`        varchar(15)  NOT NULL COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `name`        varchar(255) NOT NULL COMMENT '主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com',
    `value`       varchar(255)          DEFAULT NULL COMMENT '记录值',
    `ttl`         int(11)               DEFAULT NULL COMMENT 'TTL',
    `dns_type`    varchar(100)          DEFAULT NULL COMMENT '记录类型。<br />A | AAAA',
    `deny_global` char(1)      NOT NULL DEFAULT 'N' COMMENT '是否拒绝全局解析',
    `status`      varchar(32)  NOT NULL COMMENT '状态。<br />ENABLE-启用',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_d_host` (`host`),
    KEY `ix_d_name` (`name`)
) COMMENT ='解析记录表.';
