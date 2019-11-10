COS418作业2：Chandy-Lamport分布式快照
介绍
在此分配中，您将为分布式快照实现 Chandy-Lamport算法。您的快照算法将在令牌传递系统的顶部实现，类似于precept和Chandy-Lamport论文中介绍的系统。

该算法进行以下假设：

没有失败，所有消息都完整无缺地到达了一次
通信通道为单向且FIFO排序 (先进先出)
系统中任何两个进程之间都有一条通信路径
任何过程都可以启动快照算法
快照算法不会干扰进程的正常执行
系统中的每个进程都记录其本地状态和传入通道的状态
软件
您将在此目录下找到代码。该代码的组织方式如下：

server.go：分布式算法中的进程
simulator.go：离散时间仿真器
logger.go：一个记录器，记录由系统执行的事件（用于调试）
common.go：服务器，记录器和模拟器中使用的调试标志和常见消息类型
snapshot_test.go：您需要通过的测试
syncmap.go：线程安全映射的实现（同步的map）
queue.go：使用Go中的列表实现的简单队列接口 
test_common.go：用于测试的辅助函数
test_data /：测试用例输入和预期结果

在这些文件中，您只需要上交server.go和Simulator.go。
然而，其他一些文件还包含将是你实现或调试，有用的，
如信息调试 旗common.go和线程安全地图syncmap.go。
但是，如果您希望自己实现其功能，则不必使用提供的SyncMap。
您的任务是实现诸如TODO：IMPLEMENT ME的功能，并在必要时向周围的类添加字段。

测试中
Our grading uses the tests in snapshot_test.go provided to you. 
Test cases can be found in test_data/. To test the correctness of your code, 
simply run the following command:

  $ cd chandy-lamport/
  $ go test
  Running test '2nodes.top', '2nodes-simple.events'
  Running test '2nodes.top', '2nodes-message.events'
  Running test '3nodes.top', '3nodes-simple.events'
  Running test '3nodes.top', '3nodes-bidirectional-messages.events'
  Running test '8nodes.top', '8nodes-sequential-snapshots.events'
  Running test '8nodes.top', '8nodes-concurrent-snapshots.events'
  Running test '10nodes.top', '10nodes.events'
  PASS
  ok      _/path/to/chandy-lamport 0.012s
  
To run individual（个人） tests, 
you can look up（查阅） the test names in 
snapshot_test.go and run:

  $ go test -run 2Node
  Running test '2nodes.top', '2nodes-simple.events'
  Running test '2nodes.top', '2nodes-message.events'
  PASS
  ok      chandy-lamport  0.006s
  
Point                   Distribution
Test	                Points
2NodesSimple	        13
2NodesSingleMessage	    13
3NodesMultipleMessages	14
3NodesMultipleBidirectionalMessages	14
8NodesSequentialSnapshots	        15
8NodesConcurrentSnapshots	        15
10Nodes	                            16

Submitting Assignment
You hand in your assignment as before.

$ git commit -am "[you fill me in]"
$ git tag -a -m "i finished assignment 2" a2-handin
$ git push origin master a2-handin
回想一下，为了覆盖标签，请按如下所示使用force标志。

$ git tag -fam “我完成作业2 ” a2-handin
$ git push -f-标签
您应检查您是否能看到你的资料库GitHub的页面上最终提交并标签为这项任务。