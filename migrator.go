package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Migrator 数据迁移器
type Migrator struct {
	mysqlDB *sql.DB
	pgDB    *sql.DB
	config  *Config
}

// NewMigrator 创建新的迁移器
func NewMigrator(config *Config) (*Migrator, error) {
	// 检查数据源类型
	if config.Source.Type != "mysql" {
		return nil, fmt.Errorf("不支持的数据源类型: %s，目前只支持 mysql", config.Source.Type)
	}

	// 检查目标数据库类型
	if config.Target.Type != "postgresql" {
		return nil, fmt.Errorf("不支持的目标数据库类型: %s，目前只支持 postgresql", config.Target.Type)
	}

	// 连接MySQL
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		config.Source.MySQL.Username,
		config.Source.MySQL.Password,
		config.Source.MySQL.Host,
		config.Source.MySQL.Port,
		config.Source.MySQL.Database,
		config.Source.MySQL.Charset,
	)

	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	// 测试MySQL连接
	if err := mysqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	// 连接PostgreSQL
	pgDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Target.PostgreSQL.Host,
		config.Target.PostgreSQL.Port,
		config.Target.PostgreSQL.Username,
		config.Target.PostgreSQL.Password,
		config.Target.PostgreSQL.Database,
		config.Target.PostgreSQL.SSLMode,
	)

	pgDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %v", err)
	}

	// 测试PostgreSQL连接
	if err := pgDB.Ping(); err != nil {
		return nil, fmt.Errorf("PostgreSQL连接测试失败: %v", err)
	}

	return &Migrator{
		mysqlDB: mysqlDB,
		pgDB:    pgDB,
		config:  config,
	}, nil
}

// Close 关闭数据库连接
func (m *Migrator) Close() {
	if m.mysqlDB != nil {
		m.mysqlDB.Close()
	}
	if m.pgDB != nil {
		m.pgDB.Close()
	}
}

// MigrateTenantInfo 迁移tenant_info表
func (m *Migrator) MigrateTenantInfo() error {
	log.Println("开始迁移 tenant_info 表...")

	// 查询MySQL数据
	query := `SELECT tenant_id, tenant_name, tenant_desc FROM tenant_info`
	rows, err := m.mysqlDB.Query(query)
	if err != nil {
		return fmt.Errorf("查询MySQL tenant_info失败: %v", err)
	}
	defer rows.Close()

	// 准备PostgreSQL插入语句
	insertStmt, err := m.pgDB.Prepare(`
		INSERT INTO tenant_info (tenant_id, tenant_name, tenant_desc) 
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return fmt.Errorf("准备PostgreSQL插入语句失败: %v", err)
	}
	defer insertStmt.Close()

	// 开始事务
	tx, err := m.pgDB.Begin()
	if err != nil {
		return fmt.Errorf("开始PostgreSQL事务失败: %v", err)
	}

	count := 0
	for rows.Next() {
		var tenantID, tenantName, tenantDesc sql.NullString
		if err := rows.Scan(&tenantID, &tenantName, &tenantDesc); err != nil {
			tx.Rollback()
			return fmt.Errorf("读取MySQL行数据失败: %v", err)
		}

		// 插入到PostgreSQL
		if _, err := tx.Stmt(insertStmt).Exec(
			tenantID.String,
			tenantName.String,
			tenantDesc.String,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("插入PostgreSQL数据失败: %v", err)
		}

		count++
		if count%1000 == 0 {
			log.Printf("已迁移 %d 条 tenant_info 记录", count)
		}
	}

	if err := rows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("读取MySQL数据时发生错误: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交PostgreSQL事务失败: %v", err)
	}

	log.Printf("tenant_info 表迁移完成，共迁移 %d 条记录", count)
	return nil
}

// MigrateConfigInfo 迁移config_info表
func (m *Migrator) MigrateConfigInfo() error {
	log.Println("开始迁移 config_info 表...")

	// 查询MySQL数据
	query := `SELECT data_id, group_id, content, tenant_id, type FROM config_info`
	rows, err := m.mysqlDB.Query(query)
	if err != nil {
		return fmt.Errorf("查询MySQL config_info失败: %v", err)
	}
	defer rows.Close()

	// 准备PostgreSQL插入语句
	insertStmt, err := m.pgDB.Prepare(`
		INSERT INTO config_info (data_id, group_id, content, tenant_id, type, version, create_time) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("准备PostgreSQL插入语句失败: %v", err)
	}
	defer insertStmt.Close()

	// 开始事务
	tx, err := m.pgDB.Begin()
	if err != nil {
		return fmt.Errorf("开始PostgreSQL事务失败: %v", err)
	}

	count := 0
	currentTime := time.Now()
	for rows.Next() {
		var dataID, groupID, content, tenantID, configType sql.NullString
		if err := rows.Scan(&dataID, &groupID, &content, &tenantID, &configType); err != nil {
			tx.Rollback()
			return fmt.Errorf("读取MySQL行数据失败: %v", err)
		}

		// 插入到PostgreSQL
		if _, err := tx.Stmt(insertStmt).Exec(
			dataID.String,
			groupID.String,
			content.String,
			tenantID.String,
			configType.String,
			1, // version 固定为1
			currentTime,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("插入PostgreSQL数据失败: %v", err)
		}

		count++
		if count%1000 == 0 {
			log.Printf("已迁移 %d 条 config_info 记录", count)
		}
	}

	if err := rows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("读取MySQL数据时发生错误: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交PostgreSQL事务失败: %v", err)
	}

	log.Printf("config_info 表迁移完成，共迁移 %d 条记录", count)
	return nil
}
