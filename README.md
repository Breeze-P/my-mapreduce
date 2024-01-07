# My-MapReduce

简易版分布式MapReduce

Base on Go



### 运行方式

```bash
cd main
go run mrcoordinator.go pg-*.txt

// 新启一个命令行窗口
cd main
go build -buildmode=plugin ../mrapps/wc.go
go run mrworker.go wc.so  
```

运行于本地文件系统，中间结果存储于`/main/intermediate`运行结果于`/main/result`目录下可见。



### 目录介绍

`main`：主程序入口，包含输入文件和输出文件

`mr`：具体的coordinator和worker实现，一集rpc结构体敬意

`mrapps`：一些mapreduce应用的实现



💡

* ihash预分配reduce任务
* json格式存储中间文件
* hasTimeOut判活，无需额外开销
* 使用Unix socket作为通信方式



✅

* 完成基本业务需求：实现**分布式**mapreduce
* 通过测试脚本
* 实现基本容错：worker crash，task重分配