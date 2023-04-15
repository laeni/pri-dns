-- db_pridns.forward definition
CREATE TABLE `forward`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `client_host`        varchar(15)  NOT NULL DEFAULT '' COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `name`        varchar(255) NOT NULL COMMENT '需要转发解析的域名',
    `dns_svr`     varchar(255)          DEFAULT NULL COMMENT '转发目标DNS服务器，可以是多个，多个以逗号分割',
    `deny_global` char(1)      NOT NULL DEFAULT 'N' COMMENT '是否拒绝全局解析',
    `status`      varchar(12)  NOT NULL COMMENT '状态。<br />ENABLE-启用',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_f_clienthost` (`client_host`),
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

-- db_pridns.history_ex definition
CREATE TABLE `history_ex`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `client_host` varchar(15)  NOT NULL DEFAULT '' COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `ip_net`      varchar(255) NOT NULL COMMENT '需要排除的网段',
    `deny_global` char(1)      NOT NULL DEFAULT 'N' COMMENT '是否拒绝全局.为了简化，和 domain 表一样当 clent_host 为空时的记录对所有人生效，但是特定的某个而可以排除这种默认设置',
    `label`       varchar(100)          DEFAULT NULL COMMENT '标签/分组',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_he_clienthost` (`client_host`)
) COMMENT ='用于排除历史中的网段。由于历史数据需要按照一定规则聚合精简，所以处理后的范围可能包含一些特殊网段，比如内网等，所以这里列出的数据将在处理后生效，即精确排除这里列出的网段.';

-- db_pridns.`domain` definition
CREATE TABLE `domain`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `client_host`        varchar(15)  NOT NULL COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `name`        varchar(255) NOT NULL COMMENT '主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com',
    `value`       varchar(255)          DEFAULT NULL COMMENT '记录值',
    `ttl`         int(11)               DEFAULT NULL COMMENT 'TTL',
    `dns_type`    varchar(100)          DEFAULT NULL COMMENT '记录类型。<br />A | AAAA',
    `deny_global` char(1)      NOT NULL DEFAULT 'N' COMMENT '是否拒绝全局解析',
    `status`      varchar(32)  NOT NULL COMMENT '状态。<br />ENABLE-启用',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_d_clienthost` (`client_host`),
    KEY `ix_d_name` (`name`)
) COMMENT ='解析记录表.';
