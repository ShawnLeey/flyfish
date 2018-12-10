package flyfish

import(
	//"github.com/sniperHW/kendynet/util"
	"github.com/go-redis/redis"
	//"sync"
	//"strconv"
	"fmt"
	"strings"
	//"flyfish/conf"
	"flyfish/errcode"
	protocol "flyfish/proto"
)

type redisCmd struct {
	ctx     *processContext
	fields  []string
	ret     interface{} 
} 

type redisPipeliner struct {
	pipeLiner redis.Pipeliner
	cmds      []*redisCmd
	max       int
}

func newRedisPipeliner(max int) *redisPipeliner {
	return &redisPipeliner{
		pipeLiner : cli.Pipeline(),
		cmds      : []*redisCmd{},
		max       : max,
	} 	
}

func (this *redisPipeliner) appendIncrBy(ctx *processContext) interface{} {
	cmd := ctx.getCmd()
	keys := []string{ctx.getUniKey()}
	args := []interface{} {
		"__version__",
		ctx.fields["__version__"].GetValue(),
		cmd.incrDecr.GetName(),
		cmd.incrDecr.GetValue(),
	}
	return this.pipeLiner.Eval(strIncrBy,keys,args...)
}

func (this *redisPipeliner) appendDecrBy(ctx *processContext) interface{} {
	Debugln("appendDecrBy")
	cmd := ctx.getCmd()
	keys := []string{ctx.getUniKey()}
	args := []interface{} {
		"__version__",
		ctx.fields["__version__"].GetValue(),
		cmd.incrDecr.GetName(),
		cmd.incrDecr.GetValue(),
	}
	return this.pipeLiner.Eval(strDecrBy,keys,args...)
}

func (this *redisPipeliner) appendCompareAndSet(ctx *processContext) interface{} {
	//ARGV[1]:filed_name,ARGV[2]:old_value,ARGV[3]:new_value,ARGV[4]:__version__,ARGV[5]:__version__value
	cmd  := ctx.getCmd()
	keys := []string{ctx.getUniKey()}
	args := []interface{}{}
	args = append(args,cmd.cns.oldV.GetName(),cmd.cns.oldV.GetValue(),cmd.cns.newV.GetValue(),"__version__",ctx.fields["__version__"].GetValue())
	return this.pipeLiner.Eval(strCompareAndSet,keys,args...)
}

func (this *redisPipeliner) appendSet(ctx *processContext) interface{} {
	keys := []string{ctx.getUniKey()}
	args := []interface{}{"__version__",ctx.fields["__version__"].GetValue()}
	ARGV := []string{}
	c    := 3
	for _,v := range(ctx.fields) {
		args = append(args,v.GetName(),v.GetValue())
		ARGV = append(ARGV,fmt.Sprintf("ARGV[%d]",c),fmt.Sprintf("ARGV[%d]",c + 1))
		c += 2							
	}
	return this.pipeLiner.Eval(fmt.Sprintf(strSet,strings.Join(ARGV,",")),keys,args...)	
}

func (this *redisPipeliner) readGetResult(rcmd *redisCmd) {
	r,err1 := rcmd.ret.(*redis.SliceCmd).Result()
	if nil != err1 {
		Debugln("readGetResult error",err1)
		rcmd.ctx.errno = errcode.ERR_REDIS
	} else{
		for kk,vv := range(r) {
			name := rcmd.fields[kk]
			ckey := rcmd.ctx.getCacheKey()
			f := ckey.convertStr(name,vv.(string))
			if nil != f {
				rcmd.ctx.fields[name] = f
			}
		}
	}	
}

func (this *redisPipeliner) readSetResult(rcmd *redisCmd) {
	_,err1 := rcmd.ret.(*redis.StatusCmd).Result()
	if nil != err1 {
		Debugln("readSetResult error",err1)
		rcmd.ctx.errno = errcode.ERR_REDIS
	}	
}

func (this *redisPipeliner) readDelResult(rcmd *redisCmd) {
	r,err1 := rcmd.ret.(*redis.Cmd).Result()
	if nil != err1 {
		Debugln("cmdIncr error",err1)
		rcmd.ctx.errno = errcode.ERR_REDIS
	} else {
		if r.(string) != "ok" {
			rcmd.ctx.errno = errcode.ERR_STALE_CACHE
		} 	
	}
}

func (this *redisPipeliner) readSetScriptResult(rcmd *redisCmd) {
	r,err1 := rcmd.ret.(*redis.Cmd).Result()
	if nil != err1 {
		Debugln("cmdIncr error",err1)
		rcmd.ctx.errno = errcode.ERR_REDIS
	} else {
		cmd := rcmd.ctx.getCmd()
		if cmd.cmdType == cmdSet {
			if r.(string) != "ok" {
				rcmd.ctx.errno = errcode.ERR_STALE_CACHE
			}
		} else if cmd.cmdType == cmdIncrBy || cmd.cmdType == cmdDecrBy {
			switch r.(type) {
			case string:
				rcmd.ctx.errno = errcode.ERR_STALE_CACHE
				break
			case int64:
				rcmd.ctx.fields[cmd.incrDecr.GetName()] = protocol.PackField(cmd.incrDecr.GetName(),r.(int64))
				break
			default:
				rcmd.ctx.errno = errcode.ERR_REDIS
				break
			}
		} else if cmd.cmdType == cmdCompareAndSet || cmd.cmdType == cmdCompareAndSetNx {

			switch r.(type) {
			case string:
				rcmd.ctx.errno = errcode.ERR_STALE_CACHE
				break
			case []interface{}:
				vv := r.([]interface{})
				if vv[0].(string) == "failed" {
					rcmd.ctx.errno = errcode.ERR_NOT_EQUAL
				}
				rcmd.ctx.fields[cmd.cns.oldV.GetName()] = cmd.ckey.convertStr(cmd.cns.oldV.GetName(),vv[1].(string))
				break
			default:
				rcmd.ctx.errno = errcode.ERR_REDIS
				break
			}					
		}
	}
}


func (this *redisPipeliner) append(ctx *processContext) {
	rcmd := &redisCmd {
		ctx : ctx,
	}

	if ctx.redisFlag == redis_set || ctx.redisFlag == redis_set_only {
		Debugln("append set",ctx.redisFlag)
		rcmd.ret = this.pipeLiner.HMSet(ctx.getUniKey(),*ctx.getSetfields())
	} else if ctx.redisFlag == redis_get {
		rcmd.fields = make([]string,len(ctx.fields))
		c := 0
		for k,_ := range(ctx.fields) {
			rcmd.fields[c] = k
			c++
		}
		rcmd.ret = this.pipeLiner.HMGet(ctx.getUniKey(),rcmd.fields...)
	} else if ctx.redisFlag == redis_del {
		keys := []string{ctx.getUniKey()}
		args := []interface{}{
			"__version__",
			ctx.fields["__version__"].GetValue(),
		}
		rcmd.ret = this.pipeLiner.Eval(strDel,keys,args...)		
	} else if ctx.redisFlag == redis_set_script {
		cmdType := ctx.getCmdType()
		if cmdType == cmdCompareAndSet || cmdType == cmdCompareAndSetNx {
			rcmd.ret = this.appendCompareAndSet(ctx)
		} else if cmdType == cmdSet {
			rcmd.ret = this.appendSet(ctx)
		} else if cmdType == cmdIncrBy {
			rcmd.ret = this.appendIncrBy(ctx)
		} else if cmdType == cmdDecrBy {
			rcmd.ret = this.appendDecrBy(ctx)
		} else {
			panic("invaild cmdType")
		}
	} else {
		panic("invaild redisFlag")
	}

	this.cmds = append(this.cmds,rcmd)

	if len(this.cmds) >= this.max {
		this.exec()
	}
}



func (this *redisPipeliner) exec() {
	if len(this.cmds) == 0 {
		return
	}
	_ , err := this.pipeLiner.Exec()
	for _,v := range(this.cmds) {
		v.ctx.errno = errcode.ERR_OK
		if nil != err {
			v.ctx.errno = errcode.ERR_REDIS
			Errorln("redis exec error",err)
		} else {
			if v.ctx.redisFlag == redis_get {
				this.readGetResult(v)
			} else if v.ctx.redisFlag == redis_set || v.ctx.redisFlag == redis_set_only {
				this.readSetResult(v)
			} else if v.ctx.redisFlag == redis_del {
				this.readDelResult(v)
			} else {
				this.readSetScriptResult(v)
			}	
		}
		onRedisResp(v.ctx)
	}
	this.cmds = []*redisCmd{}
} 

