#!/bin/sh
# 单实例：第二个实例应把参数交给第一个实例后退出。
cd "$(dirname "$0")/.."
rm -f reports/single.log
./bin/validate -log reports/single.log &
APP=$!
sleep 3
OUT=$(./bin/validate open-session-demo 2>&1); CODE=$?
sleep 1
echo "second: exit=$CODE output=$OUT"
grep second-instance reports/single.log
kill $APP; wait $APP 2>/dev/null
