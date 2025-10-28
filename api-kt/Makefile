.PHONY: run build live test clean help upgradeable tidy migrate jooq

# Default environment vars
PEBBLE_PASSWORD ?= foobar
DB_PATH ?= ../data/app-dev.db

# Project settings
JAR_NAME := target/pebble-1.0.0.jar

run:
	PEBBLE_PASSWORD=$(PEBBLE_PASSWORD) SPRING_PROFILES_ACTIVE=dev ./mvnw spring-boot:run

build:
	./mvnw clean package

test:
	SPRING_PROFILES_ACTIVE=test ./mvnw test

prod-run: build
	PEBBLE_PASSWORD=$(PEBBLE_PASSWORD) SPRING_PROFILES_ACTIVE=prod java -jar $(JAR_NAME)

jooq:
	./mvnw jooq-codegen:generate

migrate:
	@echo "Using database path: $(DB_PATH)"
	./mvnw flyway:migrate -Dflyway.url=jdbc:sqlite:$(DB_PATH)

clean:
	./mvnw clean
	rm -rf $(JAR_NAME)

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
	@echo "  build - 构建"
	@echo ""
	@echo "数据库命令:"
	@echo "  jooq       - 生成 JOOQ 代码"
	@echo "  migrate    - 执行数据库迁移 (根据当前环境)"
	@echo ""
	@echo "项目维护:"
	@echo "  clean      - 清理构建产物"
	@echo ""
	@echo "环境变量:"
	@echo "  SPRING_PROFILES_ACTIVE - Spring 环境 (dev/prod/test), 默认: dev"
	@echo "  PEBBLE_PASSWORD        - 应用密码, 必须设置"

# 默认目标
.DEFAULT_GOAL := help
