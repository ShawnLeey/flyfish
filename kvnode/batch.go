package kvnode

import (
	"github.com/sniperHW/flyfish/util/fixedarray"
	"github.com/sniperHW/flyfish/util/str"
	"time"
)

var fixedArrayPool *fixedarray.FixedArrayPool
var replayOK *commitedBatchProposal = &commitedBatchProposal{}
var replaySnapshot *commitedBatchProposal = &commitedBatchProposal{}

func initFixedArrayPool(size int) {
	fixedArrayPool = fixedarray.NewPool(size)
}

//for linearizableRead

type readBatchSt struct {
	readIndex int64
	tasks     *fixedarray.FixedArray
	deadline  time.Time
}

type batchProposal struct {
	proposalStr *str.Str
	tasks       *fixedarray.FixedArray
	index       int64
}

type commitedBatchProposal struct {
	data  []byte
	tasks *fixedarray.FixedArray
}

func (this *readBatchSt) onError(err int) {
	this.tasks.ForEach(func(v interface{}) {
		v.(asynCmdTaskI).onError(err)
	})
	fixedArrayPool.Put(this.tasks)
}

func (this *readBatchSt) reply() {
	this.tasks.ForEach(func(v interface{}) {
		v.(asynCmdTaskI).reply()
		v.(asynCmdTaskI).getKV().processQueueCmd()
	})
	fixedArrayPool.Put(this.tasks)
}

func (this *batchProposal) onError(err int) {
	this.tasks.ForEach(func(v interface{}) {
		v.(asynTaskI).onError(err)
	})
	fixedArrayPool.Put(this.tasks)
	str.Put(this.proposalStr)
}

func (this *batchProposal) onPorposeTimeout() {
	this.tasks.ForEach(func(v interface{}) {
		v.(asynTaskI).onPorposeTimeout()
	})
}

func (this *commitedBatchProposal) apply(store *kvstore) {

}
