# .PHONY 声明了“伪目标”，这些目标不代表真实的文件名
.PHONY: all build build-node build-cli run test clean

# "all" 是一个常见的默认目标，它会构建所有东西
all: build

# "build" 目标现在依赖于构建节点和客户端
build: build-node build-cli

# 构建区块链节点程序
build-node:
	@echo "Building xchain node..."
	@go build -o ./bin/xchain .

# 【新增】构建命令行客户端程序
build-cli:
	@echo "Building xchain-cli..."
	@go build -o ./bin/xchain-cli ./cmd/xchain-cli

# "run" 目标现在只依赖于构建节点程序
run: build-node
	@echo "Starting xchain node network..."
	@./bin/xchain

# 测试目标保持不变
test:
	@go test -v $(if $(FILE),$(FILE),./...) $(if $(FUNC),-run $(FUNC))

# 清理目标保持不变
clean:
	@echo "Cleaning up database..."
	@rm -rf ./db