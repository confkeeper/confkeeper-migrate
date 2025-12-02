package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Source    SourceConfig    `yaml:"source"`
	Target    TargetConfig    `yaml:"target"`
	Migration MigrationConfig `yaml:"migration"`
}

// SourceConfig 数据源配置
type SourceConfig struct {
	Type  string      `yaml:"type"`
	MySQL MySQLConfig `yaml:"mysql"`
}

// TargetConfig 目标数据库配置
type TargetConfig struct {
	Type       string           `yaml:"type"`
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

// MigrationConfig 迁移配置
type MigrationConfig struct {
	BatchSize      int      `yaml:"batch_size"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	TenantNames    []string `yaml:"tenant_names"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
