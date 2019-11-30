package lamport

import (
	"log"
	"math/rand"
	"strings"
)

// Max random delay added to packet delivery 传送包的最大延迟
const maxDelay = 5

// Simulator is the entry point to the distributed snapshot application.
// 模拟器是分布式快照程序的入口点
// It is a discrete time simulator, 它是一个离散的时间模拟器
// i.e. events that happen at time t + 1 come strictly after events that happen at time t.
// 也就是说t+1时刻的时间紧跟着t时刻的时间
// At each time step, the simulator检查所有信道上排队的消息然后决定哪一个发送去目的地
// examines messages queued up across all the links in the system and decides
// which ones to deliver to the destination.

// 模拟器负责开始快照进程
// The simulator is responsible for starting the snapshot process,
// 并引导severs去相互传递令牌
// inducing servers to pass tokens to each other,
// 并且收集快照进程结束以后的快照状态
// and collecting the snapshot state after the process has terminated.

type Simulator struct {
	time           int //时间
	nextSnapshotId int
	servers        map[string]*Server // key = server ID
	logger         *Logger  //日志
	// TODO: 加了一个存储全局快照的map
	globalsnapshot map[string]SnapshotState
}
type globalsnapshot struct {
	server string
	snapstate string
	snapmap map[string]int // 分为本地local和in内向信道两种
}
func NewSimulator() *Simulator {
	return &Simulator{
		0,
		0,
		make(map[string]*Server), //使用内建函数创建一个map
		NewLogger(), //创建一个logger
		make(map[string]SnapshotState)}
}

// Return the receive time of a message after adding a random delay.
//在添加随机延迟后返回消息的接收时间
// Note:
//因为我们在每一个时间步骤只发送一个信息给sever，所以我们可能在已经收到时间步之后才收到message
// since we only deliver one message to a given server at each time step,
// the message may be received *after* the time step returned in this function.

func (sim *Simulator) GetReceiveTime() int { //返回int值
	return sim.time + 1 + rand.Intn(5)
}

//使用指定数量的启动令牌将服务器添加到此模拟器
// Add a server to this simulator with the specified number of starting tokens
func (sim *Simulator) AddServer(id string, tokens int) { // (sim *Simulator)是给Simulator类型定义了一个方法
	server := NewServer(id, tokens, sim)
	sim.servers[id] = server
}

// Add a unidirectional link（单向链接） between two servers
func (sim *Simulator) AddForwardLink(src string, dest string) {
	server1, ok1 := sim.servers[src]
	server2, ok2 := sim.servers[dest]
	if !ok1 {
		log.Fatalf("Server %v does not exist\n", src)
	}
	if !ok2 {
		log.Fatalf("Server %v does not exist\n", dest)
	}
	server1.AddOutboundLink(server2)
}

//Run an event in the system
//判断是 快照事件 还是 发送事件 还是 tick
func (sim *Simulator) InjectEvent(event interface{}) {
	switch event := event.(type) {
	case PassTokenEvent: //发送事件
		src := sim.servers[event.src] //返回服务器节点（*sever）
		src.SendTokens(event.tokens, event.dest) //把指定数量的tokens发送到dest节点
	case SnapshotEvent:  //快照事件
		sim.StartSnapshot(event.serverId) //SnapshotEvent {serverId} 中的serverid为开始快照的id
	default:
		log.Fatal("Error unknown event: ", event)
	}
}

// Advance the simulator time forward by one step,（将时间模拟器的时间向前推进一步）
// handling all send message events that expire（终止） at the new time step,
// if any.
func (sim *Simulator) Tick() {
	sim.time++
	sim.logger.NewEpoch()
	// Note: to ensure deterministic ordering of packet delivery across the servers,
	//确保服务器之间的确定性顺序
	// we must also iterate through the servers and the links in a deterministic way
	// 我们还必须以确定的方式遍历服务器和链接
	for _, serverId := range getSortedKeys(sim.servers) {
		server := sim.servers[serverId] //获取对应的服务器节点
		for _, dest := range getSortedKeys(server.outboundLinks) { //获得该服务器的外向信道的目的地
			link := server.outboundLinks[dest] //获得该服务器到dest的外向信道
			// Deliver at most one packet per server at each time step to
			// establish total ordering of packet delivery to each server
			//在每个时间步骤中，每个服务器最多交付一个包，以确定向每个服务器交付包的总顺序
			if !link.events.Empty() { //信道里的事件不为空的情况下
				e := link.events.Peek().(SendMessageEvent) //获取信道里的发送事件
				if e.receiveTime <= sim.time {
					link.events.Pop() //信道中的sendmessevent的信息弹出队列
					sim.logger.RecordEvent(
						sim.servers[e.dest],
						ReceivedMessageEvent{e.src, e.dest, e.message})
					sim.servers[e.dest].HandlePacket(e.src, e.message) //接收服务器.HandlePacket(源服务器，传输的信息)
					break
				}
			}
		}
	}
}

// Start a new snapshot process at the specified（在指定的） server
//在serverid指定的服务器上开始一个节点
func (sim *Simulator) StartSnapshot(serverId string) {
	snapshotId := sim.nextSnapshotId
	sim.nextSnapshotId++
	sim.logger.RecordEvent(sim.servers[serverId], StartSnapshot{serverId, snapshotId})
	//todo:IMPLEMENT ME
	serversrc := sim.servers[serverId] 	//获取到开始快照的服务器
	serversrc.StartSnapshot(snapshotId) //开始一个快照
}

// Callback for servers to notify（通知） the simulator that
// the snapshot process has completed（完成） on a particular（特定的） server
//通知模拟器id为serverid的服务器已经完成快照了
func (sim *Simulator) NotifySnapshotComplete(serverId string, snapshotId int) {
	sim.logger.RecordEvent(sim.servers[serverId], EndSnapshot{serverId, snapshotId})
	// TODO: 服务器通知模拟器本服务器已经完成了快照
	sim.globalsnapshot[serverId] = sim.servers[serverId].snapshot
	sim.servers[serverId].snapshotstate = "complete"

}


// Collect and merge（合并） snapshot state from all the servers.
// This function blocks(阻碍) until the snapshot process has completed on all servers.
//收集快照的函数
func (sim *Simulator) CollectSnapshot(snapshotId int) *SnapshotState {
	snap := SnapshotState{snapshotId, make(map[string]int), make([]*SnapshotMessage, 0)}
	// TODO: IMPLEMENT ME
	bool := false
	count := 0
	for _,serverid := range getSortedKeys(sim.servers){
		if(strings.EqualFold(sim.servers[serverid].snapshotstate,"complete")){
			count += 1
		}
	}
	if(count == len(sim.servers)){
		bool = true
	}
	if(bool){
		for _,serverid := range getSortedKeys(sim.servers){
			snap.tokens[serverid] = sim.servers[serverid].snapshot.tokens["local"] + sim.servers[serverid].snapshot.tokens["in"]
		}
	}

	return &snap
}
