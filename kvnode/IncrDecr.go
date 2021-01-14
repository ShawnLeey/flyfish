package kvnode

import (
	pb "github.com/golang/protobuf/proto"
	"github.com/sniperHW/flyfish/errcode"
	"github.com/sniperHW/flyfish/net"
	"github.com/sniperHW/flyfish/proto"
	"time"
)

type asynCmdTaskIncrDecr struct {
	*asynCmdTaskBase
}

func (this *asynCmdTaskIncrDecr) onSqlResp(errno int32) {
	this.asynCmdTaskBase.onSqlResp(errno)
	cmd := this.commands[0].(*cmdIncrDecr)

	if !cmd.checkVersion(this.version) {
		this.errno = errcode.ERR_VERSION_MISMATCH
	} else {
		if errno == errcode.ERR_RECORD_NOTEXIST {
			this.fields = map[string]*proto.Field{}
			fillDefaultValue(cmd.getKV().meta, &this.fields)
			this.sqlFlag = sql_insert_update
			this.version = time.Now().UnixNano()
		} else {
			this.sqlFlag = sql_update
		}

		this.errno = errcode.ERR_OK
		oldV := this.fields[cmd.field.GetName()]
		var newV *proto.Field

		if cmd.isIncr {
			newV = proto.PackField(oldV.GetName(), oldV.GetInt()+cmd.field.GetInt())
		} else {
			newV = proto.PackField(oldV.GetName(), oldV.GetInt()-cmd.field.GetInt())
		}

		this.fields[oldV.GetName()] = newV
		this.version = increaseVersion(this.version)
	}

	if this.errno != errcode.ERR_OK {
		this.reply()
		this.getKV().store.issueAddkv(this)
	} else {
		this.getKV().store.issueUpdate(this)
	}
}

func newAsynCmdTaskIncrDecr() *asynCmdTaskIncrDecr {
	return &asynCmdTaskIncrDecr{
		asynCmdTaskBase: &asynCmdTaskBase{
			commands: []commandI{},
		},
	}
}

type cmdIncrDecr struct {
	*commandBase
	field  *proto.Field
	isIncr bool
}

func (this *cmdIncrDecr) reply(errCode int32, fields map[string]*proto.Field, version int64) {
	this.replyer.reply(this, errCode, fields, version)
}

func (this *cmdIncrDecr) makeResponse(errCode int32, fields map[string]*proto.Field, version int64) *net.Message {

	var pbdata pb.Message
	var field *proto.Field

	if errcode.ERR_OK == errCode {
		field = fields[this.field.GetName()]
	}

	if this.isIncr {
		pbdata = &proto.IncrByResp{
			Version: version,
			Field:   field,
		}
	} else {
		pbdata = &proto.DecrByResp{
			Version: version,
			Field:   field,
		}
	}

	return net.NewMessage(net.CommonHead{
		Seqno:   this.replyer.seqno,
		ErrCode: errCode,
	}, pbdata)

}

func (this *cmdIncrDecr) prepare(t asynCmdTaskI) (asynCmdTaskI, bool) {

	task, ok := t.(*asynCmdTaskIncrDecr)

	if t != nil && !ok {
		//不同类命令不能合并
		return t, false
	}

	if t != nil && this.version != nil {
		//同类命令但需要检查版本号，不能跟之前的命令合并
		return t, false
	}

	kv := this.kv
	status := kv.getStatus()

	if t != nil && status == cache_new {
		//需要从数据库加载，不合并
		return t, false
	}

	if status != cache_new && !this.checkVersion(kv.version) {
		this.reply(errcode.ERR_VERSION_MISMATCH, nil, kv.version)
		return t, true
	}

	if nil == t {
		task = newAsynCmdTaskIncrDecr()
		task.fields = map[string]*proto.Field{}
		if status == cache_missing {
			fillDefaultValue(kv.meta, &task.fields)
			task.sqlFlag = sql_insert_update
		} else if status == cache_ok {
			task.sqlFlag = sql_update
		}
	}

	if status != cache_new {
		var newV *proto.Field
		var oldV *proto.Field

		oldV = task.fields[this.field.GetName()]

		if nil == oldV {
			oldV = kv.fields[this.field.GetName()]
		}

		if nil == oldV {
			v := kv.meta.GetDefaultV(this.field.GetName())
			if nil == v {
				this.reply(errcode.ERR_INVAILD_FIELD, nil, kv.version)
				return t, true
			} else {
				oldV = proto.PackField(this.field.GetName(), kv.meta.GetDefaultV(this.field.GetName()))
			}
		}

		if this.isIncr {
			newV = proto.PackField(oldV.GetName(), oldV.GetInt()+this.field.GetInt())
		} else {
			newV = proto.PackField(oldV.GetName(), oldV.GetInt()-this.field.GetInt())
		}

		task.fields[oldV.GetName()] = newV

		if nil == t {
			if status == cache_missing {
				task.version = time.Now().UnixNano()
			} else {
				task.version = increaseVersion(kv.version)
			}
		}
	}

	task.commands = append(task.commands, this)

	return task, true

}
