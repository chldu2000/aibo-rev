#!/bin/sh
# Q87：Aibo 进程被 kill -9 后，各类子进程是否随之结束。
cd "$(dirname "$0")/.."
rm -f reports/children.json
./bin/validate -children reports/children.json -log reports/children.log &
APP=$!
for i in $(seq 1 50); do [ -s reports/children.json ] && break; sleep 0.2; done
sleep 0.5
kill -9 $APP
sleep 1.5
python3 - <<'PY'
import json, os
d = json.load(open('reports/children.json'))
import time
def alive(pid):
    try:
        os.kill(pid, 0); return True
    except ProcessLookupError:
        return False
res = []
for c in d['children']:
    c['pgid'] = os.getpgid(c['pid']) if alive(c['pid']) else None
    res.append({**c, 'aliveAfterAppKilled': alive(c['pid'])})
# 模拟“重启时清理残留”：按记录的进程组结束
for c in res:
    if c['aliveAfterAppKilled'] and c['pgid']:
        try: os.killpg(c['pgid'], 9)
        except ProcessLookupError: pass
time.sleep(0.3)
for c in res:
    c['aliveAfterGroupKill'] = alive(c['pid'])
print(json.dumps(res, ensure_ascii=False))
json.dump(res, open('reports/children-result.json', 'w'), ensure_ascii=False, indent=2)
PY
