package sqlnode

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/sniperHW/flyfish/errcode"
	"github.com/sniperHW/flyfish/net"
	"time"
)
import "github.com/sniperHW/flyfish/proto"

type sqlTaskGet struct {
	sqlCombinableTaskBase
	getAll    bool
	getFields []string
	//version   int64
	//getWithVersion bool
	//versionOp      string
	fields map[string]*proto.Field
}

func (t *sqlTaskGet) combine(c cmd) bool {
	cmd, ok := c.(*cmdGet)
	if !ok {
		return false
	}

	if c.uniKey() != t.uniKey {
		return false
	}

	//if cmd.version == nil && t.getWithVersion {
	//	t.getWithVersion = false
	//} else if cmd.version != nil && t.getWithVersion && *cmd.version < t.version {
	//	t.version = *cmd.version
	//	t.versionOp = ">="
	//}

	if cmd.getAll {
		t.getAll = true
		t.getFields = nil
	} else if !t.getAll {
		for _, v := range cmd.fields {
			if _, ok := t.fields[v]; !ok {
				t.getFields = append(t.getFields, v)
				t.fields[v] = nil
			}
		}
	}

	t.addCmd(c)

	return true
}

func (t *sqlTaskGet) do(db *sqlx.DB) {
	tableMeta := getDBMeta().getTableMeta(t.table)

	var (
		queryFields          []string
		queryFieldReceivers  []interface{}
		queryFieldConverters []fieldConverter
		queryFieldCount      int
		getFieldOffset       int
		versionIndex         int
		errCode              int32
		version              int64
		sqlStr               = getStr()
	)

	if t.getAll {
		queryFields = tableMeta.getAllFields()
		queryFieldReceivers = tableMeta.getAllFieldReceivers()
		queryFieldConverters = tableMeta.getAllFieldConverter()
		queryFieldCount = len(queryFields)
		getFieldOffset = 2
		versionIndex = versionFieldIndex

		appendSelectAllSqlStr(sqlStr, tableMeta, t.key, nil)
	} else {
		getFieldCount := len(t.getFields)
		queryFieldCount = getFieldCount + 1
		queryFields = make([]string, queryFieldCount)
		queryFieldReceivers = make([]interface{}, queryFieldCount)
		queryFieldConverters = make([]fieldConverter, queryFieldCount)

		queryFields[0] = versionFieldName
		queryFieldReceivers[0] = versionFieldMeta.getReceiver()
		queryFieldConverters[0] = versionFieldMeta.getConverter()
		versionIndex = 0

		getFieldOffset = 1
		for i := 0; i < getFieldCount; i++ {
			fieldMeta := tableMeta.getFieldMeta(t.getFields[i])
			queryFields[i+getFieldOffset] = t.getFields[i]
			queryFieldReceivers[i+getFieldOffset] = fieldMeta.getReceiver()
			queryFieldConverters[i+getFieldOffset] = fieldMeta.getConverter()
		}

		sqlStr.AppendString("select ")
		sqlStr.AppendString(queryFields[0])
		for i := 1; i < queryFieldCount; i++ {
			sqlStr.AppendString(",")
			sqlStr.AppendString(queryFields[i])
		}
		sqlStr.AppendString(" from ").AppendString(t.table).AppendString(" where ").AppendString(keyFieldName).AppendString("='").AppendString(t.key).AppendString("';")
	}

	s := sqlStr.ToString()
	start := time.Now()
	row := db.QueryRowx(s)
	getLogger().Debugf("task-get: table(%s) key(%s): query:\"%s\" cost:%.3fs.", t.table, t.key, s, time.Now().Sub(start).Seconds())

	if err := row.Scan(queryFieldReceivers...); err == sql.ErrNoRows {
		errCode = errcode.ERR_RECORD_NOTEXIST
		//if t.getWithVersion {
		//	t.errCode = errcode.ERR_RECORD_UNCHANGE
		//} else {
		//	t.errCode = errcode.ERR_RECORD_NOTEXIST
		//}
	} else if err != nil {
		getLogger().Errorf("task-get table(%s) key(%s): %s.", t.table, t.key, err)
		errCode = errcode.ERR_SQLERROR
	} else {
		errCode = errcode.ERR_OK
		version = queryFieldConverters[versionIndex](queryFieldReceivers[versionIndex]).(int64)

		for i := getFieldOffset; i < queryFieldCount; i++ {
			fieldName := queryFields[i]
			t.fields[fieldName] = proto.PackField(fieldName, queryFieldConverters[i](queryFieldReceivers[i]))
		}
	}

	putStr(sqlStr)
	t.reply(errCode, version, t.fields)
}

type cmdGet struct {
	cmdBase
	getAll  bool
	fields  []string
	version *int64
}

//func (c *cmdGet) getReq() *proto.GetReq {
//	return c.msg.GetData().(*proto.GetReq)
//}

func (c *cmdGet) makeSqlTask() sqlTask {
	tableFieldCount := getDBMeta().getTableMeta(c.table).getFieldCount()

	task := &sqlTaskGet{
		sqlCombinableTaskBase: newSqlCombinableTaskBase(c.uKey, c.table, c.key),
		fields:                make(map[string]*proto.Field, tableFieldCount),
	}

	if c.getAll {
		task.getAll = true
	} else {
		task.getFields = make([]string, 0, tableFieldCount)
		for _, v := range c.fields {
			task.getFields = append(task.getFields, v)
			task.fields[v] = proto.PackField(v, nil)
		}
	}

	//pVersion := c.version
	//if pVersion != nil {
	//	task.version = *pVersion
	//	task.getWithVersion = true
	//	task.versionOp = "!="
	//} else {
	//	task.version = 0
	//	task.getWithVersion = false
	//}

	task.addCmd(c)

	return task
}

func (c *cmdGet) reply(errCode int32, version int64, fields map[string]*proto.Field) {
	if !c.isResponseTimeout() {
		var resp = &proto.GetResp{
			Version: version,
			Fields:  nil,
		}

		if errCode == errcode.ERR_OK {
			if c.version != nil && version == *c.version {
				errCode = errcode.ERR_RECORD_UNCHANGE
			} else {
				if c.getAll {
					resp.Fields = make([]*proto.Field, len(fields))
					i := 0
					for _, v := range fields {
						resp.Fields[i] = v
						i++
					}
				} else {
					resp.Fields = make([]*proto.Field, len(c.fields))
					for i, v := range c.fields {
						if f := fields[v]; f == nil {
							getLogger().Errorf("get table(%s) key(%s) lost field(%s).", c.table, c.key, v)
						} else {
							resp.Fields[i] = f
						}
					}
				}
			}
		}

		_ = c.conn.sendMessage(net.NewMessage(
			net.CommonHead{
				Seqno:   c.sqNo,
				ErrCode: errCode,
			},
			resp,
		))
	}
}

func onGet(conn *cliConn, msg *net.Message) {
	req := msg.GetData().(*proto.GetReq)

	head := msg.GetHead()

	table, key := head.SplitUniKey()

	tableMeta := getDBMeta().getTableMeta(table)

	if tableMeta == nil {
		getLogger().Errorf("get table(%s) key(%s): table not exist.", table, key)
		_ = conn.sendMessage(newMessage(head.Seqno, errcode.ERR_INVAILD_TABLE, &proto.GetResp{}))
		return
	}

	if !req.GetAll() {
		if len(req.GetFields()) == 0 {
			getLogger().Errorf("get table(%s) key(%s): no fields.", table, key)
			_ = conn.sendMessage(newMessage(head.Seqno, errcode.ERR_MISSING_FIELDS, &proto.GetResp{}))
			return
		}

		if b, i := tableMeta.checkFieldNames(req.GetFields()); !b {
			getLogger().Errorf("get table(%s) key(%s): invalid field(%s).", table, key, req.GetFields()[i])
			_ = conn.sendMessage(newMessage(head.Seqno, errcode.ERR_INVAILD_FIELD, &proto.GetResp{}))
			return
		}
	}

	processDeadline, respDeadline := getDeadline(head.Timeout)

	cmd := &cmdGet{
		cmdBase: newCmdBase(conn, head.Seqno, head.UniKey, table, key, processDeadline, respDeadline),
		getAll:  req.GetAll(),
		fields:  req.GetFields(),
		version: req.Version,
	}

	pushCmd(cmd)
}
