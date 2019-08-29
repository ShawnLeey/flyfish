package server

import (
	"encoding/binary"
	"fmt"
	"github.com/sniperHW/flyfish/conf"
	"github.com/sniperHW/flyfish/errcode"
	"github.com/sniperHW/flyfish/proto"
	"hash/crc64"
	"os"
	"sync/atomic"
	"time"
)

const (
	binlog_none     = 0
	binlog_snapshot = 1
	binlog_update   = 2
	binlog_delete   = 3
	binlog_kick     = 4
)

type binlogSt struct {
	binlogStr        *str
	ctxs             *ctxArray
	cacheBinlogCount int32
}

var (
	fileCounter  int64
	checkSumSize = 8
	crc64Table   *crc64.Table
	binlogSuffix = ".bin"
	tmpFileName  string
)

func onWriteFileError(err error) {
	//写文件错误可能是因为磁盘满导致，所以先删除预留文件，释放磁盘空间用来写日志
	os.Remove(tmpFileName)
	Errorln("onWriteFileError", err)
	os.Exit(1)
}

func (this *kvstore) startSnapshot() {

	if this.make_snapshot {
		return
	}

	config := conf.GetConfig()

	this.make_snapshot = true

	this.backFilePath = this.filePath
	this.f.Close()

	fileIndex := atomic.AddInt64(&fileCounter, 1)
	os.MkdirAll(config.BinlogDir, os.ModePerm)
	path := fmt.Sprintf("%s/%s_%d%s", config.BinlogDir, config.BinlogPrefix, fileIndex, binlogSuffix)

	f, err := os.Create(path)
	if err != nil {
		Fatalln("create backfile failed", path, err)
	}

	this.binlogStr = strGet()

	this.binlogCount = 0
	this.fileSize = 0

	this.f = f
	this.filePath = path

	kv := []*kv{}

	for _, v := range this.kv {
		v.mtx.Lock()
		if v.status == cache_ok || v.status == cache_missing {
			v.snapshoted = false
			kv = append(kv, v)
		}
		v.mtx.Unlock()
	}

	go func() {
		beg := time.Now()
		Infoln("start snapshot")
		c := 0
		i := 0
		for _, v := range kv {
			this.mtx.Lock()
			v.mtx.Lock()
			if (v.status == cache_ok || v.status == cache_missing) && !v.snapshoted {
				c++
				v.snapshoted = true
				this.writeBinlog(binlog_snapshot, v.uniKey, v.values, v.version)

			}
			v.make_snapshot = false
			v.mtx.Unlock()
			this.mtx.Unlock()
			i++
			if i%100 == 0 {
				time.Sleep(time.Millisecond * 10)
			}
		}

		//移除backfile
		os.Remove(this.backFilePath)

		this.mtx.Lock()
		this.make_snapshot = false
		this.mtx.Unlock()
		Infoln("snapshot ok", time.Now().Sub(beg), c)
	}()
}

func (this *kvstore) flushBinlog(binlogStr *str, ctxs *ctxArray, cacheBinlogCount int32) {
	this.mtx.Lock()

	beg := time.Now()

	config := conf.GetConfig()

	if nil == this.f {

		fileIndex := atomic.AddInt64(&fileCounter, 1)

		os.MkdirAll(config.BinlogDir, os.ModePerm)
		path := fmt.Sprintf("%s/%s_%d%s", config.BinlogDir, config.BinlogPrefix, fileIndex, binlogSuffix)

		f, err := os.Create(path)
		if err != nil {
			Fatalln("create backfile failed", path, err)
			return
		}

		this.f = f
		this.filePath = path
	}

	head := make([]byte, 4+checkSumSize)
	checkSum := crc64.Checksum(binlogStr.bytes(), crc64Table)
	binary.BigEndian.PutUint32(head[0:4], uint32(binlogStr.dataLen()))
	binary.BigEndian.PutUint64(head[4:], uint64(checkSum))

	this.fileSize += binlogStr.dataLen() + len(head)

	this.mtx.Unlock()

	if _, err := this.f.Write(head); nil != err {
		onWriteFileError(err)
	}

	if _, err := this.f.Write(binlogStr.bytes()); nil != err {
		onWriteFileError(err)
	}

	if err := this.f.Sync(); nil != err {
		onWriteFileError(err)
	}

	this.mtx.Lock()

	if this.binlogCount >= config.MaxBinlogCount || this.fileSize >= int(config.MaxBinlogFileSize) {
		this.startSnapshot()
	}

	Debugln("flush time:", time.Now().Sub(beg), cacheBinlogCount)

	this.mtx.Unlock()

	if nil != ctxs {
		/*   binlog持久化完成
		 *   响应客户端并把更新提交到内存
		 */
		for i := 0; i < ctxs.count; i++ {
			v := ctxs.ctxs[i]
			v.reply(errcode.ERR_OK, v.fields, v.version)
			ckey := v.getCacheKey()
			ckey.mtx.Lock()

			ckey.snapshoted = true

			if v.writeBackFlag == write_back_insert || v.writeBackFlag == write_back_update || v.writeBackFlag == write_back_insert_update {
				ckey.setValueNoLock(v)
				ckey.setOKNoLock(v.version)
			} else {
				ckey.setMissingNoLock()
			}

			ckey.sqlFlag = v.writeBackFlag

			if !ckey.writeBackLocked {
				ckey.writeBackLocked = true
				pushSqlWriteReq(ckey)
			}
			ckey.mtx.Unlock()
		}

		for i := 0; i < ctxs.count; i++ {
			v := ctxs.ctxs[i]
			v.getCacheKey().processQueueCmd()
		}
		ctxArrayPut(ctxs)
	}
	strPut(binlogStr)
}

func (this *kvstore) tryFlushBinlog() {

	if this.cacheBinlogCount > 0 {

		config := conf.GetConfig()

		if this.cacheBinlogCount >= int32(config.FlushCount) || this.binlogStr.dataLen() >= config.FlushSize || time.Now().After(this.nextFlush) {

			cacheBinlogCount := this.cacheBinlogCount

			this.cacheBinlogCount = 0

			binlogStr := this.binlogStr
			ctxs := this.ctxs

			this.binlogStr = nil
			this.ctxs = nil

			this.binlogQueue.AddNoWait(&binlogSt{
				binlogStr:        binlogStr,
				ctxs:             ctxs,
				cacheBinlogCount: cacheBinlogCount,
			})
		}
	}
}

func (this *kvstore) writeBinlog(tt int, unikey string, fields map[string]*proto.Field, version int64) {

	if nil == this.binlogStr {
		this.binlogStr = strGet()
	}

	this.binlogCount++
	this.cacheBinlogCount++

	if this.cacheBinlogCount == 1 {
		this.nextFlush = time.Now().Add(time.Millisecond * time.Duration(conf.GetConfig().FlushInterval))
	}

	//写操作码1byte
	this.binlogStr.appendByte(byte(tt))
	//写unikey
	this.binlogStr.appendInt32(int32(len(unikey)))
	this.binlogStr.append(unikey)
	//写version
	this.binlogStr.appendInt64(version)
	if tt == binlog_snapshot || tt == binlog_update {
		pos := this.binlogStr.len
		this.binlogStr.appendInt32(int32(0))
		if nil != fields {
			c := 0
			for n, v := range fields {
				if n != "__version__" {
					c++
					this.binlogStr.appendField(v)
				}
			}
			if c > 0 {
				binary.BigEndian.PutUint32(this.binlogStr.data[pos:pos+4], uint32(c))
			}
		}
	} else {
		this.binlogStr.appendInt32(int32(0))
	}
}

/*
func (this *kvstore) checkCacheKey(ckey *kv, ctx *cmdContext) bool {
	ckey.mtx.Lock()

	gotErr := false

	switch ckey.sqlFlag {
	case write_back_none:
		ckey.sqlFlag = ctx.writeBackFlag
	case write_back_insert, write_back_insert_update:
		if ctx.writeBackFlag == write_back_update {
			ckey.sqlFlag = write_back_insert_update
		} else if ctx.writeBackFlag == write_back_delete {
			ckey.sqlFlag = write_back_delete
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	case write_back_delete:
		if ctx.writeBackFlag == write_back_insert {
			ckey.sqlFlag = write_back_insert
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	case write_back_update:
		if ctx.writeBackFlag == write_back_update {
			ckey.sqlFlag = write_back_update
		} else if ctx.writeBackFlag == write_back_delete {
			ckey.sqlFlag = write_back_delete
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	default:
		gotErr = true
		Errorln("invaild ctx.writeBackFlag")
	}

	if gotErr {
		ckey.mtx.Unlock()
		ctx.reply(errcode.ERR_ERROR, nil, -1)
		ckey.processQueueCmd()
		return false
	} else {

		cmdType := ctx.getCmdType()

		if cmdType != cmdDel && nil == ckey.values {
			ckey.setDefaultValueNoLock()
		}

		switch cmdType {
		case cmdIncrBy, cmdDecrBy:
			cmd := ctx.getCmd()
			var newV *proto.Field
			oldV := ckey.values[cmd.incrDecr.GetName()]
			if cmdType == cmdIncrBy {
				newV = proto.PackField(cmd.incrDecr.GetName(), oldV.GetInt()+cmd.incrDecr.GetInt())
			} else {
				newV = proto.PackField(cmd.incrDecr.GetName(), oldV.GetInt()-cmd.incrDecr.GetInt())
			}
			ckey.modifyFields[newV.GetName()] = true
			ckey.values[newV.GetName()] = newV
			ctx.fields[newV.GetName()] = newV
			ckey.setOKNoLock(ckey.version + 1)
		case cmdDel:
			ckey.setMissingNoLock()
		case cmdSet, cmdSetNx, cmdCompareAndSet, cmdCompareAndSetNx:
			for k, v := range ctx.fields {
				if k != "__version__" {
					ckey.values[k] = v
					ckey.modifyFields[k] = true
				}
			}
			ckey.setOKNoLock(ckey.version + 1)
		}

		ctx.version = ckey.version

		ckey.mtx.Unlock()
		return true
	}
}
*/

func fillDefaultValue(meta *table_meta, ctx *cmdContext) {
	for _, v := range meta.fieldMetas {
		defaultV := proto.PackField(v.name, v.defaultV)
		if _, ok := ctx.fields[v.name]; !ok {
			ctx.fields[v.name] = defaultV
		}
	}
}

func (this *kvstore) checkCacheKey(ckey *kv, ctx *cmdContext) (bool, int) {
	ckey.mtx.Lock()

	gotErr := false
	binop := binlog_none
	sqlFlag := write_back_none

	switch ckey.sqlFlag {
	case write_back_none:
		sqlFlag = ctx.writeBackFlag
	case write_back_insert, write_back_insert_update:
		if ctx.writeBackFlag == write_back_update {
			sqlFlag = write_back_insert_update
		} else if ctx.writeBackFlag == write_back_delete {
			sqlFlag = write_back_delete
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	case write_back_delete:
		if ctx.writeBackFlag == write_back_insert {
			sqlFlag = write_back_insert
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	case write_back_update:
		if ctx.writeBackFlag == write_back_update {
			sqlFlag = write_back_update
		} else if ctx.writeBackFlag == write_back_delete {
			sqlFlag = write_back_delete
		} else {
			gotErr = true
			Errorln("invaild ctx.writeBackFlag")
		}
	default:
		gotErr = true
		Errorln("invaild ctx.writeBackFlag")
	}

	if gotErr {
		ckey.mtx.Unlock()
		ctx.reply(errcode.ERR_ERROR, nil, -1)
		ckey.processQueueCmd()
		return false, binop
	} else {

		cmdType := ctx.getCmdType()

		if cmdType != cmdDel && nil == ckey.values {
			fillDefaultValue(ckey.getMeta(), ctx)
		}

		switch cmdType {
		case cmdIncrBy, cmdDecrBy:
			cmd := ctx.getCmd()
			var newV *proto.Field
			var oldV *proto.Field
			if nil != ckey.values {
				oldV = ckey.values[cmd.incrDecr.GetName()]
			} else {
				oldV = ctx.fields[cmd.incrDecr.GetName()]
			}
			if cmdType == cmdIncrBy {
				newV = proto.PackField(cmd.incrDecr.GetName(), oldV.GetInt()+cmd.incrDecr.GetInt())
			} else {
				newV = proto.PackField(cmd.incrDecr.GetName(), oldV.GetInt()-cmd.incrDecr.GetInt())
			}
			ctx.fields[newV.GetName()] = newV
			ctx.version = ckey.version + 1
		case cmdDel:
			ctx.version = 0
		case cmdSet, cmdSetNx, cmdCompareAndSet, cmdCompareAndSetNx:
			ctx.version = ckey.version + 1
		}

		switch sqlFlag {
		case write_back_delete:
			if ckey.snapshoted {
				binop = binlog_delete
			} else {
				binop = binlog_snapshot
			}
		case write_back_insert, write_back_insert_update:
			binop = binlog_snapshot
		case write_back_update:
			if ckey.snapshoted {
				binop = binlog_update
			} else {
				binop = binlog_snapshot
			}
		}

		if binop == binlog_snapshot && sqlFlag != write_back_delete && ckey.values != nil {
			for k, v := range ckey.values {
				if _, ok := ctx.fields[k]; !ok {
					ctx.fields[k] = v
				}
			}
		}

		ckey.mtx.Unlock()
		ctx.writeBackFlag = sqlFlag
		return true, binop
	}
}

func (this *kvstore) processUpdate(ctx *cmdContext) {

	if ctx.writeBackFlag == write_back_none {
		panic("ctx.writeBackFlag == write_back_none")
	}

	ckey := ctx.getCacheKey()

	//if !this.checkCacheKey(ckey, ctx) {
	//	return
	//}

	ok, binop := this.checkCacheKey(ckey, ctx)

	if !ok {
		return
	}

	this.mtx.Lock()
	//ckey.mtx.Lock()

	if nil == this.ctxs {
		this.ctxs = ctxArrayGet()
	}

	this.ctxs.append(ctx)

	/*switch ckey.sqlFlag {
	case write_back_delete:
		if ckey.snapshoted {
			this.writeBinlog(binlog_delete, ckey.uniKey, nil, 0)
		} else {
			ckey.snapshoted = true
			this.writeBinlog(binlog_snapshot, ckey.uniKey, nil, 0)
		}
	case write_back_insert, write_back_insert_update:
		ckey.snapshoted = true
		this.writeBinlog(binlog_snapshot, ckey.uniKey, ckey.values, ckey.version)
	case write_back_update:
		if ckey.snapshoted {
			this.writeBinlog(binlog_update, ckey.uniKey, ctx.fields, ckey.version)
		} else {
			ckey.snapshoted = true
			this.writeBinlog(binlog_snapshot, ckey.uniKey, ckey.values, ckey.version)
		}
	}*/

	//ckey.mtx.Unlock()

	if len(ctx.fields) == 0 || ctx.version == 0 {
		panic("len(ctx.fields == 0) || ctx.version == 0")
	}

	this.writeBinlog(binop, ckey.uniKey, ctx.fields, ctx.version)

	this.tryFlushBinlog()

	this.mtx.Unlock()

}

func init() {
	crc64Table = crc64.MakeTable(crc64.ISO)
}