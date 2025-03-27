# redis实现优先队列

```sh
# 清空现有队列（可选）
redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL $QUEUE_KEY

# 添加元素（带优先级）
echo "Adding elements to priority queue..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 2 "task1"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 1 "task2"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 3 "task3"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $QUEUE_KEY 1 "task4"

# 查看队列状态
echo "Current queue state:"
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZRANGE $QUEUE_KEY 0 -1 WITHSCORES

# 批量弹出（使用 ZPOPMIN，Redis 5.0+ 支持）
echo "Popping 2 elements from the head of the queue..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT ZPOPMIN $QUEUE_KEY 2
```
