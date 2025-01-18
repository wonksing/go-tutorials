# Valkey

```bash
docker run -it --net wonknet --ip 172.18.0.103 -p 6379:6379 --name valkey -d valkey/valkey valkey-server --loglevel debug
```

```bash
valkey-cli

DEL reserve:1 reserve:2 reserve:3 reserve:4 reserve:5

ZRANGE reserve:1 0 -1
ZRANGE reserve:2 0 -1
ZRANGE reserve:3 0 -1
ZRANGE reserve:4 0 -1
ZRANGE reserve:5 0 -1

```
