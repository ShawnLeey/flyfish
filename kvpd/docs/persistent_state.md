# pd的持久化信息

* 1 节点配置信息

节点由一个唯一id确定合法范围是1-9999。节点启动后需要连接pd,向pd索取配置信息用以启动节点。节点配置信息包括:

	servicePort int
	raftPort    int
	ip          string //配置的ip地址
	regions     []int  //node上运行的shard

* 2 region信息

每个region有唯一的单独整数标识，根据配置，region被分配到多个节点上运行。所有这些region的副本构成一个raft group。

	regionNo int
	leader   int               //leader节点id
	nodes    map[int]*nodeConf //region关联的node
	slots    Bitmap            //region分配的slot


每个region上分配了第一数量的slot,slot是数据管理的最小单元。整个数据空间被划分为大小为65536的slot。根据key的hash值，可以将数据映射到slot中。


数据查找关系

1 (hash(key) % 65536) + 1 计算出slot

2 根据映射表，查找到slot归属的region

3 取得region leader

4 将请求发往leader


* 3 迁移事务 

将选的的slot从一个region迁移到另外一个region


slot迁移流程

迁移slot n

参与成员

转出shard leader (后面称为leader1）
转入shard leader (后面称为leader2)
pd


安全性保证

保证leader1将所有与slot n相关的kv回写到数据库后，才允许leader2接管slot n.
如果leader1所在的raft group整体故障(无法选出新leader继续执行)，slot n将处于不可用状态(无法确定
slot n相关的数据全部回写到数据库，如果此时允许其它节点接管slot n可能会读取到旧的数据从而违背一致性)。


迁移协议

pd为迁移处理创建一个全局单调递增的迁移事务id n.用此id向leader1发出prepare请求。

leader1接收到之后向raft提交一条准备执行事务id n的日志。日志复制成功后响应pd。

pd接收到响应后将事务状态设置为执行，并通知leader1开始执行事务。


leader1接收到pd发过来的执行命令后执行以下处理流程。

1 向raft log提交一条执行事务的日志
2 将slot n设置为迁移状态，在此状态下所有与slot n相关的kv请求立即返回错误码:迁移中
3 对于slot n相关的kv,如果有正在排队的请求，全部返回错误码:迁移中
4 对于slot n相关的kv,执行kick(kick保证了正确回写数据库)
5 向raft log提交一条事务完成日志
6 向pd通告事务执行完毕（需要一直执行直到pd确认）




pd接收到事务执行完毕的通告后，

1 提交一条raft log,标记leader1执行完毕,响应leader1。
2 向leader2通知接管slot n.接收到leader2的响应后更新路由表，标记事务完成。
3 向接入节点推送最新路由表。


迁移事务中的故障处理

1 leader1可能故障，选出新leader继续执行迁移事务。

leader1在事务的每个阶段都要把事务状态提交到raft日志，以允许新leader可以继续执行尚未完成的事务。


2 prepare响应超时

如果prepare响应超时，pd认为本次迁移事务失败。

leader1在接收到提交请求时，判断提交的事务id是否与之前prepare的一致，如果不一致拒绝执行。

对于leader如果事务尚未进入执行阶段，新的更大的事务id将替换当前事务id(出现场景：leader1向pd返回prepare响应，消息丢失，pd超时事务失败，一段时间之后使用新的事务id重新请求执行迁移事务).


3 pd故障

1 pd在发出prepare之后故障。

事务状态提交到raft log,新leader发现有迁移事务处于prepare时，向leader1再次发出相同的prepare请求。
leader1发现id一致，响应prepare请求，事务恢复正常执行流程。


2 pd接收到leader1事务执行完毕通告，执行步骤1后故障。

由新的pd leader继续执行后面2，3两步。



路由表项目

slot -> shard 映射表

shard leader(初始时为空，当leader产生或变更后通告pd)


配置信息

节点条目

节点id
对外服务端口
ip地址(如果连上来的节点ip与配置不符，拒绝接入)
shard(1-N个)


节点启动后连接pd leader,通过自身ip以及节点id向pd leader索要配置信息。
根据配置中的shard创建kvstore。使用配置端口启动对外服务。



接入节点处理kv请求(已经缓存正确路由信息)

根据key的hash值计算归属slot.

通过 slot -> shard映射表获取正确shard.

根据shard leader表获得 目标kvnode。向目标kvnode转发请求。