package conf

var RedisProcessPoolSize    = int(5)
var SqlLoadPoolSize         = int(5)
var SqlUpdatePoolSize       = int(5)
var RedisPipelineSize       = int(50)
var SqlLoadPipeLineSize     = int(50)
var SqlUpdatePipeLineSize   = int(500)
var SqlEventQueueSize       = int(10000)
var RedisEventQueueSize     = int(5000)
var WriteBackEventQueueSize = int(50000)
var MainEventQueueSize      = int(50000)
var MaxPacketSize           = uint64(1024*1024*4)
var WriteBackDelay          = int64(5)