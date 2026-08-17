CREATE DATABASE demo
CHARACTER SET utf8mb4
COLLATE utf8mb4_0900_ai_ci;

CREATE TABLE `user`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `company_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司id',
    `department_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '部门id',
    `name`          VARCHAR(255) NOT NULL DEFAULT '' COMMENT '姓名',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME              DEFAULT NULL COMMENT '删除时间',
    `created_by`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人id',
    `updated_by`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新人id',
    `deleted_by`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除人id',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT = '用户表';

CREATE TABLE `user_login_log`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `user_id`    bigint unsigned NOT NULL COMMENT '用户ID',
    `login_ip`   varchar(32)       DEFAULT NULL COMMENT '登录IP地址',
    `user_agent` varchar(512)      DEFAULT NULL COMMENT '用户代理信息',
    `login_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` datetime          DEFAULT NULL COMMENT '删除时间',
    `created_by` bigint unsigned NOT NULL COMMENT '创建人ID',
    `updated_by` bigint unsigned NOT NULL COMMENT '更新人ID',
    `deleted_by` bigint unsigned DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    KEY          `idx_user_id` (`user_id`),
    KEY          `idx_login_time` (`login_time`),
    KEY          `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户登录日志表';