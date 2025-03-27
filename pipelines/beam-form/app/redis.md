# redis实现优先队列

## 优先队列操作
### 清空队列
```sh
redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL $QUEUE_KEY
```

### 添加元素（带优先级）
```sh
echo "Adding elements to priority queue..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 2 "task1"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 1 "task2"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 3 "task3"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 1 "task4"
```
### 查看队列状态
```sh
echo "Current queue state:"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZRANGE $QUEUE_KEY 0 -1 WITHSCORES
```

### 批量弹出（使用 ZPOPMIN，Redis 5.0+ 支持）
```sh
echo "Popping 2 elements from the head of the queue..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZPOPMIN $QUEUE_KEY 2
```

### 应用内常用操作
### 查队列中元素(无本地redis-cli)
```sh
docker exec -t redis-server redis-cli ZRANGE QUEUE_HOSTS 0 -1 WITHSCORES
```

### 插入队列的示例(bash)
```sh
hostname=n01.p419
timestamp=$(date +%s%6N)
priority=0.1
redis-cli ZADD QUEUE_HOSTS $priority "$hostname:$timestamp"
```
