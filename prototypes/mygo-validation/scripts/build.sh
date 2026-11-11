#!/bin/sh
# 构建默认界面原型并嵌入验证程序。
set -e
cd "$(dirname "$0")/.."
(cd ../stream-composer && pnpm build >/dev/null)
rm -rf web/app && cp -R ../stream-composer/dist web/app
cp panels/panel.html web/lab/panel.html
CGO_ENABLED=0 go build -o bin/validate .
echo "built bin/validate"
