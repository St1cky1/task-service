#!/bin/bash

# Скрипт для генерации Go кода из proto файлов
# Proto файлы находятся в proto/, сгенерированный код идет в proto/pb/

set -e

# Переменные
PROTO_DIR="proto"
PB_DIR="proto/pb"
PROTOC_BIN="./protoc/bin/protoc"
GOOGLE_API_DIR="third_party"

# Проверяем наличие protoc
if [ ! -f "$PROTOC_BIN" ]; then
    echo "❌ protoc not found at $PROTOC_BIN"
    exit 1
fi

# Создаем директорию pb если её нет
mkdir -p "$PB_DIR"

echo "🔄 Generating proto files from $PROTO_DIR..."

# Генерируем proto файлы с путем относительно текущей директории
$PROTOC_BIN \
    -I "$PROTO_DIR" \
    -I "protoc/include" \
    -I "$GOOGLE_API_DIR" \
    --go_out="$PB_DIR" \
    --go-grpc_out="$PB_DIR" \
    --grpc-gateway_out="$PB_DIR" \
    --grpc-gateway_opt=paths=source_relative \
    "$PROTO_DIR"/task_service.proto \
    "$PROTO_DIR"/user_service.proto

echo "✅ Proto files generated successfully in $PB_DIR"

# Список сгенерированных файлов
if ls "$PB_DIR"/*.pb.go 1> /dev/null 2>&1; then
    echo "📋 Generated files:"
    ls -lh "$PB_DIR"/*.pb.go | awk '{print "   " $9}'
fi