-- ============================================================================
-- 初始化脚本（PostgreSQL）
--
-- 本项目以 PostgreSQL 为样板库。首次使用：
--   createdb demo
--   psql -d demo -f backend/scripts/sql/table_init.sql
--
-- 注意：user 是 PostgreSQL 保留字，凡出现在 SQL 中都必须写成 "user"（双引号）。
-- Go 侧不受此限制：golib/gormdao（v1.32.26 起）会把表名与列名交给 GORM 按当前方言引用，
-- 所以 model.TableNameUser 是**不带引号**的逻辑表名，传给 gormdao.NewDao 即可，
-- 生成的 SQL 形如 SELECT count(*) FROM "user" WHERE "user"."deleted_at" IS NULL。
--
-- 其余表（core_file、core_file_upload）由 golib/filestore 在应用启动时 AutoMigrate 创建。
--
-- 字段注释规范（见 docs/agents/backend-conventions.md）：每个字段都必须有行内 `--` 注释，
-- 并在建表后用 `COMMENT ON COLUMN` 把同一份说明写入库元数据（psql 里 \d+ 表名 可见）。
-- 两者必须保持一致，新增/改名字段时同步修改。
-- ============================================================================

-- 数据库若尚未创建，可先执行（已存在时会报错，可忽略）：
-- CREATE DATABASE demo;

-- ----------------------------------------------------------------------------
-- 用户表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS "user"
(
    id            BIGSERIAL    PRIMARY KEY,               -- 自增ID
    company_id    BIGINT       NOT NULL DEFAULT 0,        -- 公司id
    department_id BIGINT       NOT NULL DEFAULT 0,        -- 部门id
    name          VARCHAR(255) NOT NULL DEFAULT '',       -- 姓名
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 创建时间
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 更新时间（由 GORM 维护）
    deleted_at    TIMESTAMPTZ           DEFAULT NULL,     -- 删除时间（软删除）
    created_by    BIGINT       NOT NULL DEFAULT 0,        -- 创建人id
    updated_by    BIGINT       NOT NULL DEFAULT 0,        -- 更新人id
    deleted_by    BIGINT       NOT NULL DEFAULT 0         -- 删除人id
);

CREATE INDEX IF NOT EXISTS idx_user_deleted_at ON "user" (deleted_at);

COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user".id IS '自增ID';
COMMENT ON COLUMN "user".company_id IS '公司id';
COMMENT ON COLUMN "user".department_id IS '部门id';
COMMENT ON COLUMN "user".name IS '姓名';
COMMENT ON COLUMN "user".created_at IS '创建时间';
COMMENT ON COLUMN "user".updated_at IS '更新时间（由 GORM 维护）';
COMMENT ON COLUMN "user".deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN "user".created_by IS '创建人id';
COMMENT ON COLUMN "user".updated_by IS '更新人id';
COMMENT ON COLUMN "user".deleted_by IS '删除人id';

-- ----------------------------------------------------------------------------
-- 用户登录日志表
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_login_log
(
    id         BIGSERIAL    PRIMARY KEY,               -- 自增ID
    user_id    BIGINT       NOT NULL,                  -- 用户ID
    login_ip   VARCHAR(32)           DEFAULT NULL,     -- 登录IP地址
    user_agent VARCHAR(512)          DEFAULT NULL,     -- 用户代理信息
    login_time TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 登录时间
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 创建时间
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 更新时间（由 GORM 维护）
    deleted_at TIMESTAMPTZ           DEFAULT NULL,     -- 删除时间
    created_by BIGINT       NOT NULL,                  -- 创建人ID
    updated_by BIGINT       NOT NULL,                  -- 更新人ID
    deleted_by BIGINT                DEFAULT NULL      -- 删除人ID
);

CREATE INDEX IF NOT EXISTS idx_user_login_log_user_id ON user_login_log (user_id);
CREATE INDEX IF NOT EXISTS idx_user_login_log_login_time ON user_login_log (login_time);
CREATE INDEX IF NOT EXISTS idx_user_login_log_deleted_at ON user_login_log (deleted_at);

COMMENT ON TABLE user_login_log IS '用户登录日志表';
COMMENT ON COLUMN user_login_log.id IS '自增ID';
COMMENT ON COLUMN user_login_log.user_id IS '用户ID';
COMMENT ON COLUMN user_login_log.login_ip IS '登录IP地址';
COMMENT ON COLUMN user_login_log.user_agent IS '用户代理信息';
COMMENT ON COLUMN user_login_log.login_time IS '登录时间';
COMMENT ON COLUMN user_login_log.created_at IS '创建时间';
COMMENT ON COLUMN user_login_log.updated_at IS '更新时间（由 GORM 维护）';
COMMENT ON COLUMN user_login_log.deleted_at IS '删除时间';
COMMENT ON COLUMN user_login_log.created_by IS '创建人ID';
COMMENT ON COLUMN user_login_log.updated_by IS '更新人ID';
COMMENT ON COLUMN user_login_log.deleted_by IS '删除人ID';
