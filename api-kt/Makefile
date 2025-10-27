.PHONY: run build live test clean help upgradeable tidy migrate jooq

# 默认环境变量
SPRING_PROFILES_ACTIVE ?= dev
PEBBLE_PASSWORD ?= foobar

# 项目配置
JAR_NAME := target/pebble-1.0.0.jar
DB_PATH := ../data/app-$(SPRING_PROFILES_ACTIVE).db

## 开发相关命令
run:
	PEBBLE_PASSWORD=$(PEBBLE_PASSWORD) ./mvnw spring-boot:run

build:
	./mvnw clean package

test:
	SPRING_PROFILES_ACTIVE=test ./mvnw test

## 生产相关命令
prod-build:
	SPRING_PROFILES_ACTIVE=prod ./mvnw clean package

prod-run: prod-build
	PEBBLE_PASSWORD=$(PEBBLE_PASSWORD) SPRING_PROFILES_ACTIVE=prod java -jar $(JAR_NAME)

## 数据库相关命令
jooq:
	./mvnw jooq-codegen:generate

migrate:
	./mvnw flyway:migrate -Dflyway.url=jdbc:sqlite:$(DB_PATH)

migrate-dev:
	./mvnw flyway:migrate -Dflyway.url=jdbc:sqlite:../data/app-dev.db

migrate-prod:
	./mvnw flyway:migrate -Dflyway.url=jdbc:sqlite:../data/app-prod.db

## 项目维护命令
clean:
	./mvnw clean
	rm -rf $(JAR_NAME)

dependencies:
	./mvnw versions:display-dependency-updates

## 帮助信息
help:
	@echo "Spring Boot Kotlin 项目 Makefile 命令:"
	@echo ""
	@echo "开发命令:"
	@echo "  run        - 运行开发服务器 (需要设置 PEBBLE_PASSWORD 环境变量)"
	@echo "  live       - 运行开发服务器 (支持热重载)"
	@echo "  build      - 构建项目"
	@echo "  test       - 运行测试"
	@echo ""
	@echo "生产命令:"
	@echo "  prod-build - 构建生产版本"
	@echo "  prod-run   - 运行生产版本"
	@echo ""
	@echo "数据库命令:"
	@echo "  jooq       - 生成 JOOQ 代码"
	@echo "  migrate    - 执行数据库迁移 (根据当前环境)"
	@echo "  migrate-dev - 执行开发环境数据库迁移"
	@echo "  migrate-prod - 执行生产环境数据库迁移"
	@echo ""
	@echo "项目维护:"
	@echo "  clean      - 清理构建产物"
	@echo "  upgradeable - 检查可升级的依赖"
	@echo "  tidy       - 整理和格式化代码"
	@echo "  dependencies - 显示依赖更新信息"
	@echo ""
	@echo "环境变量:"
	@echo "  SPRING_PROFILES_ACTIVE - Spring 环境 (dev/prod/test), 默认: dev"
	@echo "  PEBBLE_PASSWORD        - 应用密码, 必须设置"

# 默认目标
.DEFAULT_GOAL := help
