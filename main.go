package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("请指定配置文件路径，使用 --config 参数")
	}

	// 读取配置
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 创建迁移器
	migrator, err := NewMigrator(config)
	if err != nil {
		log.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 执行迁移
	fmt.Println("开始数据迁移...")

	// 迁移 tenant_info 表
	fmt.Println("正在迁移 tenant_info 表...")
	if err := migrator.MigrateTenantInfo(); err != nil {
		log.Fatalf("迁移 tenant_info 表失败: %v", err)
	}
	fmt.Println("tenant_info 表迁移完成")

	// 迁移 config_info 表
	fmt.Println("正在迁移 config_info 表...")
	if err := migrator.MigrateConfigInfo(); err != nil {
		log.Fatalf("迁移 config_info 表失败: %v", err)
	}
	fmt.Println("config_info 表迁移完成")

	fmt.Println("所有表迁移完成！")
}
