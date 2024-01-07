package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

type EmptyArgs struct {
}

type EmptyReply struct {
}

type GetIDReply struct {
	Id int
}

type GetTaskArgs struct {
	Id int
}

type GetTaskReply struct {
	T         string   // 任务类型
	TaskId    int      // 任务ID
	FileNames []string // 输入文件位置
	NReduce   int      // Reduce任务数
}

type DoneMapArgs struct {
	Id        int
	FileNames []string // 中间文件存储地址
}

type DoneReduceArgs struct {
	Id int
}

type KeepAliveArgs struct {
	Id int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
