-- db_pridns.forward definition
CREATE TABLE `forward`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    `name`        varchar(255) NOT NULL COMMENT '需要转发解析的域名',
    `dns`         varchar(255) NOT NULL COMMENT '转发目标DNS服务器，可以是多个，多个以逗号分割',
    `bind`        varchar(15)           DEFAULT NULL COMMENT '生效范围。可选值：全局或某个具体Ip(如果为空则表示全局)',
    `enabled`     varchar(12)  NOT NULL COMMENT '是否启用',
    `history`     varchar(255)          DEFAULT NULL COMMENT '该域名对应的解析历史。可能需要导出使用',
    PRIMARY KEY (`id`),
    KEY `ix_f_name` (`name`)
) COMMENT ='转发配置.';

-- db_pridns.`domain` definition
CREATE TABLE `domain`
(
    `id`          bigint(20)   NOT NULL AUTO_INCREMENT,
    `host`        varchar(15)  NOT NULL COMMENT '客户端地址（生效范围）。<br />如果全局生效，则该字段为空。',
    `name`        varchar(255) NOT NULL COMMENT '主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com',
    `value`       varchar(255)          DEFAULT NULL COMMENT '记录值',
    `ttl`         int(11)               DEFAULT NULL COMMENT 'TTL',
    `status`      varchar(32)  NOT NULL COMMENT '状态。<br />ENABLE-启用',
    `type`        varchar(256)          DEFAULT NULL COMMENT '记录类型。带"NO_"前缀的表示用于禁用全局解析。<br />A / NO_ALL / NO_A',
    `priority`    int(11)      NOT NULL DEFAULT 1 COMMENT '优先级。值越小优先级越高。',
    `create_time` datetime     NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `update_time` datetime     NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_d_host` (`host`) USING BTREE,
    KEY `ix_d_name` (`name`) USING BTREE
) COMMENT ='解析记录表.';
