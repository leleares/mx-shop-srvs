# mx-shop-srvs 
慕学商城微服务service层
- user_srv 用户微服务
- goods_srv 商品微服务

### 分布式锁
锁是操作系统给予的资源调度的能力，对同一台服务器上的并发请求使用锁可以很好的解决资源竞争问题。但是一旦统一服务部署为多台服务器，各个服务器之间都有各自的一把锁，无法对同一资源进行很好的并发资源控制，分布式锁就是解决这个问题的，可以理解为抽象于各个服务器之上的一把锁，各个服务共享。

常见的分布式锁实现方案
- 基于Mysql的悲观锁、乐观锁
- 基于redis的分布式锁
- 基于zookeeper的分布式锁

MySql悲观锁：
我很悲观，总是认为数据资源会有人来抢，所以趁早把资源锁住。（MySql语法支持）
注意：悲观锁在MySql中属于行锁，即会锁住一条记录，但前提是for update的查询条件字段是有索引的，如果没有索引，就不是锁一条记录了，而是会升级成表锁。
```sql
START TRANSACTION;

SELECT stock
FROM inventory
WHERE good_id = 1 //注意good_id一定为加了索引的字段
FOR UPDATE; // 这里就把 good_id = 1的这条记录上锁了

UPDATE inventory
SET stock = stock - 5
WHERE good_id = 1; 

COMMIT; // 释放锁
```

MySql乐观锁：
我很乐观，等真正提交时再检查有没有冲突。（通过加字段来模拟实现）