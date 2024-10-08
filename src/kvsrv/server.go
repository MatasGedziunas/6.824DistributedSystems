package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu        sync.Mutex
	db        map[string]string
	callsDone map[string]int
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	lastOpId, exists := kv.callsDone[args.ClientId]
	if !exists || args.OperationId != lastOpId {
		reply.Value = kv.db[args.Key]
		log.Printf("GET applied (duplicate) ; Client=%s, Operation=%d", args.ClientId, args.OperationId)
	} else {
		log.Printf("GETskipped (duplicate) ; Client=%s, Operation=%d", args.ClientId, args.OperationId)
	}
	kv.callsDone[args.ClientId] = args.OperationId
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	lastOpId, exists := kv.callsDone[args.ClientId]
	// log.Printf("Put: ClientId=%s, OperationId=%d, LastOpId=%d, Exists=%v, Key=%s, Value=%s", args.ClientId, args.OperationId, lastOpId, exists, args.Key, args.Value)
	reply.Value = kv.db[args.Key]
	if !exists || args.OperationId != lastOpId {
		kv.db[args.Key] = args.Value
		log.Printf("Put applied: Key=%s, Value=%s, Client=%s, Operation=%d", args.Key, args.Value, args.ClientId, args.OperationId)
	} else {
		log.Printf("Put skipped (duplicate) ; Client=%s, Operation=%d", args.ClientId, args.OperationId)
	}
	kv.callsDone[args.ClientId] = args.OperationId
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	lastOpId, exists := kv.callsDone[args.ClientId]
	// log.Printf("Append: ClientId=%s, OperationId=%d, LastOpId=%d, Exists=%v", args.ClientId, args.OperationId, lastOpId, exists)
	reply.Value = kv.db[args.Key]
	if !exists || args.OperationId != lastOpId {
		kv.db[args.Key] += args.Value
		kv.callsDone[args.ClientId] = args.OperationId
		log.Printf("Append applied: ClientId=%s, OperationId=%d, LastOpId=%d, Exists=%v", args.ClientId, args.OperationId, lastOpId, exists)
	} else {
		log.Printf("Append skipped: ClientId=%s, OperationId=%d, LastOpId=%d, Exists=%v", args.ClientId, args.OperationId, lastOpId, exists)
	}
	kv.callsDone[args.ClientId] = args.OperationId
}

func StartKVServer() *KVServer {
	kv := new(KVServer)
	kv.db = make(map[string]string)
	kv.callsDone = make(map[string]int)
	return kv
}
