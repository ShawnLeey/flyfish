package flyfish

import (
	"fmt"
	codec "flyfish/codec"
	message "flyfish/proto"
	"flyfish/errcode"
	"github.com/sniperHW/kendynet"
	"github.com/golang/protobuf/proto"
)

type GetReplyer struct {
	seqno      int64
	session    kendynet.StreamSession
	context    *cmdContext
}

func (this *GetReplyer) reply(errCode int32,fields map[string]field,version ...int64) {
	resp := &message.GetResp{
		Seqno : proto.Int64(this.seqno),
		ErrCode : proto.Int32(errCode),
	}

	if len(version) > 0 {
		resp.Version = proto.Int64(version[0])
	}

	if errcode.ERR_OK == errCode {
		for _,field := range(this.context.fields) {
			if v,ok := fields[field.name];ok {
				resp.Fields = append(resp.Fields,message.PackField(field.name,v.value))	
			}
		}
	}

	Debugln("GetReply",this.context.uniKey,resp)

	err := this.session.Send(resp)
	if nil != err {
		//记录日志
		Debugln("send GetReply error",this.context.uniKey,resp,err)
	}
}


type GetAllReplyer struct {
	seqno      int64
	session    kendynet.StreamSession
	context    *cmdContext
}

func (this *GetAllReplyer) reply(errCode int32,fields map[string]field,version ...int64) {
	resp := &message.GetallResp{
		Seqno : proto.Int64(this.seqno),
		ErrCode : proto.Int32(errCode),
	}

	if len(version) > 0 {
		resp.Version = proto.Int64(version[0])
	}

	if errcode.ERR_OK == errCode {
		for _,field := range(this.context.fields) {
			if v,ok := fields[field.name];ok {
				resp.Fields = append(resp.Fields,message.PackField(field.name,v.value))	
			}
		}
	}

	Debugln("GetAllReply",this.context.uniKey,resp)

	err := this.session.Send(resp)
	if nil != err {
		//记录日志
		Debugln("send GetAllReply error",this.context.uniKey,resp,err)
	}
}

func getAll(session kendynet.StreamSession,msg *codec.Message) {
	req := msg.GetData().(*message.GetallReq)
	errno := errcode.ERR_OK

	Debugln("getAll",req)

	if "" == req.GetTable() {
		errno = errcode.ERR_CMD_MISSING_TABLE
	}

	if "" == req.GetKey() {
		errno = errcode.ERR_CMD_MISSING_KEY
	}	

	meta := GetMetaByTable(req.GetTable())

	if nil == meta {
		errno = errcode.ERR_INVAILD_TABLE
	}

	if errcode.ERR_OK != errno {
		resp := &message.GetallResp{
			Seqno : proto.Int64(req.GetSeqno()),
			ErrCode : proto.Int32(errno),
		}
		err := session.Send(resp)
		if nil != err {
			//记录日志
		}				
		return
	}

	context := &cmdContext{
		cmdType   : cmdGet,
		key       : req.GetKey(),
		table     : req.GetTable(),
		uniKey    : fmt.Sprintf("%s:%s",req.GetTable(),req.GetKey()),
		fields    : []field{},
	}

	context.rpyer = &GetAllReplyer{
		seqno : req.GetSeqno(),
		session : session,
		context : context,		
	}

	for _,name := range(meta.queryMeta.field_names) {
		context.fields = append(context.fields,field{
			name : name,
		})		
	}
	pushCmdContext(context)
}

func get(session kendynet.StreamSession,msg *codec.Message) {
	req := msg.GetData().(*message.GetReq)
	errno := errcode.ERR_OK

	Debugln("get",req)

	if "" == req.GetTable() {
		errno = errcode.ERR_CMD_MISSING_TABLE
	}

	if "" == req.GetKey() {
		errno = errcode.ERR_CMD_MISSING_KEY
	}

	if errcode.ERR_OK != errno {
		resp := &message.GetResp{
			Seqno : proto.Int64(req.GetSeqno()),
			ErrCode : proto.Int32(errno),
		}
		err := session.Send(resp)
		if nil != err {
			//记录日志
		}				
		return
	}
	
	context := &cmdContext{
		cmdType   : cmdGet,
		key       : req.GetKey(),
		table     : req.GetTable(),
		uniKey    : fmt.Sprintf("%s:%s",req.GetTable(),req.GetKey()),
		fields    : []field{},
	}

	context.rpyer = &GetReplyer{
		seqno : req.GetSeqno(),
		session : session,
		context : context,		
	}

	for _,v := range(req.GetFields()) {
		context.fields = append(context.fields,field{
			name : v,
		})
	}
	pushCmdContext(context)
}



