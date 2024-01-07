package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// TODO 其实id没用，都是靠在数组中的下标标识object的
type MapTask struct {
	id       int
	status   int // 0: free, 1: working, 2: done
	fileName string

	startTime time.Time
	workerID  int
}

type ReduceTask struct {
	id        int
	status    int
	fileNames []string

	startTime time.Time
	workerID  int
}

type WorkerInstance struct {
	id     int
	status int // 0: free, 1: working, 2: died
}

type Coordinator struct {
	// Your definitions here.

	mu         sync.Mutex
	mapDone    bool
	reduceDone bool

	mapTasks    []MapTask
	reduceTasks []ReduceTask
	workers     []WorkerInstance
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) GetID(args *EmptyArgs, reply *GetIDReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := len(c.workers)
	reply.Id = id
	c.workers = append(c.workers, WorkerInstance{id, 0})
	fmt.Println("recept a worker ", id)
	return nil
}

// Function to check if a task has timed out
func hasTimeOut(startTime time.Time) bool {
	timeoutDuration := 10 * time.Second
	return time.Since(startTime) > timeoutDuration
}

func (c *Coordinator) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.mapDone { // 分配map任务
		reply.T = "Map"
		for i, m := range c.mapTasks {
			if m.status == 1 {
				if hasTimeOut(m.startTime) { // 超时重新分配任务
					c.workers[c.mapTasks[i].workerID].status = 2 // 标记死亡
					c.workers[args.Id].status = 1

					c.mapTasks[i].workerID = args.Id
					c.mapTasks[i].startTime = time.Now()

					reply.TaskId = m.id
					reply.FileNames = []string{m.fileName}
					reply.NReduce = len(c.reduceTasks)
					fmt.Println("A map task has been distributed again, ", m.id)
					return nil
				}
			}
			if m.status == 0 {
				c.workers[args.Id].status = 1

				c.mapTasks[i].status = 1 // 要更改原数据
				c.mapTasks[i].workerID = args.Id
				c.mapTasks[i].startTime = time.Now()

				reply.TaskId = m.id
				reply.FileNames = []string{m.fileName}
				reply.NReduce = len(c.reduceTasks)
				fmt.Println("A map task has been distributed, ", m.id)
				return nil
			}
		}
		reply.T = "Dial" // 无任务可分配
		return nil
	}
	if !c.reduceDone { // 分配reduce任务
		reply.T = "Reduce"
		for i, r := range c.reduceTasks {
			if r.status == 1 {
				if hasTimeOut(r.startTime) { // 超时重新分配任务
					c.workers[c.reduceTasks[i].workerID].status = 2 // 标记死亡
					c.workers[args.Id].status = 1

					c.reduceTasks[i].workerID = args.Id
					c.reduceTasks[i].startTime = time.Now()

					reply.TaskId = r.id
					reply.FileNames = r.fileNames
					reply.NReduce = len(c.reduceTasks)
					fmt.Println("A reduce task has been distributed again, ", r.id)
					return nil
				}
			}
			if r.status == 0 {
				c.workers[args.Id].status = 1

				c.reduceTasks[i].status = 1 // 要更改原数据
				c.reduceTasks[i].workerID = args.Id
				c.reduceTasks[i].startTime = time.Now()

				reply.TaskId = r.id
				reply.FileNames = r.fileNames
				reply.NReduce = len(c.reduceTasks)
				fmt.Println("A reduce task has been distributed ", r.id)
				return nil
			}
		}
		reply.T = "Dial" // 无任务可分配
		return nil
	}
	reply.T = "Done" // 结束
	return nil
}

func (c *Coordinator) DoneMap(args *DoneMapArgs, reply *EmptyReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := args.Id
	outputFileNames := args.FileNames
	c.mapTasks[id].status = 2
	for i, file := range outputFileNames {
		c.reduceTasks[i].fileNames = append(c.reduceTasks[i].fileNames, file)
	}
	fmt.Printf("Map Task %v Over\n", id)
	for _, m := range c.mapTasks {
		if m.status != 2 {
			return nil
		}
	}
	c.mapDone = true
	fmt.Println("Map Tasks All Over")
	return nil
}

func (c *Coordinator) DoneReduce(args *DoneReduceArgs, reply *EmptyReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := args.Id
	c.reduceTasks[id].status = 2
	fmt.Printf("Reduce Task %v Over\n", id)
	for _, r := range c.reduceTasks {
		if r.status != 2 {
			return nil
		}
	}
	c.reduceDone = true
	fmt.Println("All Tasks Over")
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	fmt.Println("handle a request")
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("Listen on unix socket ", sockname)
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := c.mapDone && c.reduceDone

	// for _, r := range c.reduceTasks {
	// 	if r.status != 2 {
	// 		return false
	// 	}
	// }
	// ret = true

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.

	// 设置MapTasks数为输入文件数目
	nMap := len(files)

	// 初始化MapTasks
	c.mapTasks = make([]MapTask, nMap)
	for i := 0; i < nMap; i++ {
		c.mapTasks[i].id = i
		c.mapTasks[i].fileName = files[i]
	}

	// 初始化ReduceTasks
	c.reduceTasks = make([]ReduceTask, nReduce)
	for i := 0; i < nReduce; i++ {
		c.reduceTasks[i].id = i
	}

	c.server()
	return &c
}
