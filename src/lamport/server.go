package lamport

import (
	"log"
	"strings"
)

// The main participant（参与者） of the distributed snapshot protocol（协议）.
// Servers exchange（交换） token messages and marker messages among each other.
// Token messages represent the transfer（调动） of tokens from one server to another.
// Marker messages represent the progress（进度） of the snapshot process.
// The bulk（大多数） of the distributed protocol is implemented（实现） in `HandlePacket（处理数据包）` and `StartSnapshot（开始快照）`.

type Server struct {
	Id            string
	Tokens        int
	sim           *Simulator
	outboundLinks map[string]*Link // key = link.dest 输出信道
	inboundLinks  map[string]*Link // key = link.src 输入信道
	// TODO: ADD MORE FIELDS HERE
	localtokens int
	snapshotstate string //start notstart complete 三种状态
	snapshot SnapshotState
}

type SnapState struct {
	tokens int
	State string
}
// A unidirectional（单向） communication通信 channel between two servers
// Each link contains an event queue（事件队列） (as opposed to（而不是） a packet queue)
type Link struct {
	src    string
	dest   string
	events *Queue
}

func NewServer(id string, tokens int, sim *Simulator) *Server {
	return &Server{
		id,
		tokens,
		sim,
		make(map[string]*Link),
		make(map[string]*Link),
		0,
		"notstart",
		SnapshotState{sim.nextSnapshotId,make(map[string]int),make([]*SnapshotMessage,0)}}
}

// Add a unidirectional（单向） link to the destination（目的地）server
func (server *Server) AddOutboundLink(dest *Server) {
	if server == dest {
		return
	}
	l := Link{server.Id, dest.Id, NewQueue()}
	server.outboundLinks[dest.Id] = &l
	dest.inboundLinks[server.Id] = &l
}

// Send a message on all of the server's outbound links
// 向所有相邻的外向信道发送信息
func (server *Server) SendToNeighbors(message interface{}) {
	for _, serverId := range getSortedKeys(server.outboundLinks) {
		link := server.outboundLinks[serverId]
		server.sim.logger.RecordEvent(
			server,
			SentMessageEvent{server.Id, link.dest, message})
		link.events.Push(SendMessageEvent{
			server.Id,
			link.dest,
			message,
			server.sim.GetReceiveTime()})
	}
}

// 把指定数量的tokens发送到指定的服务器节点
func (server *Server) SendTokens(numTokens int, dest string) {
	if server.Tokens < numTokens {
		log.Fatalf(
			"{Server %v} a想要发送 %v tokens 给 {sever %v}，但是他只有 %v 个节点\n",
					server.Id, numTokens, dest,server.Tokens)
	}
	log.Printf("{Server %v} a想要发送 %v tokens 给 {sever %v}，{%v节点} 有 %v 个tokens\n",
		server.Id, numTokens, dest,server.Id,server.Tokens)

	message := TokenMessage{numTokens}
	server.sim.logger.RecordEvent(server, SentMessageEvent{server.Id, dest, message})
	// Update local state before sending the tokens
	server.Tokens -= numTokens // 减去源服务器要发送的 numtokens
	link, ok := server.outboundLinks[dest] //获得跟dest服务器连接的外向信道
	if !ok {
		log.Fatalf("未知的 dest ID %v from 源 server %v\n", dest, server.Id)
	}
	link.events.Push(
		SendMessageEvent{
		server.Id,
		dest,
		message,
		server.sim.GetReceiveTime()})
}

// Callback（回收信号） for when a message is received on this server.
// When the snapshot algorithm completes on this server,
// this function should notify（通知） the simulator by calling `sim.NotifySnapshotComplete`
func (server *Server) HandlePacket(src string, message interface{}) { //接收服务器 .HandlePacket(源服务器，传输的信息)
	//TODO:IMPLEMENT ME
	switch message := message.(type) {
	case TokenMessage: //普通token消息
		if strings.EqualFold(server.snapshotstate, "start") {
			server.snapshot.tokens["in"] += message.numTokens
		}
		server.Tokens += message.numTokens
		log.Printf(" ------------------------------------ %v 服务器收到了tokens的数 %v --------------------------- ",server.Id,message.numTokens)
	case MarkerMessage:    //marker消息 ，向相邻的信道发送
		if(strings.EqualFold(server.snapshotstate,"start")){
			//todo：
			// 1 .把这个marker接收存起来
			// 2. 判断这个服务器有几个外向信道，并且是否都从这些信道里收到了marker
			// 	2.1 如果是本服务器收到了不是所有信道都收到了marker信息
			// 	2.2 如果服务器的全部外向信道都收到了marker消息，的时候调用 sim.NotifySnapshotComplete 方法通知模拟器该服务器已经完成了快照
			for _,link := range server.outboundLinks{
				src := link.src
				dest := link.dest
				event := link.events

				if(!event.Empty()){
					if(len(server.snapshot.messages) == len(server.outboundLinks)){
						//这种情况说明所有外向信道都收到了marker
						server.sim.NotifySnapshotComplete(server.Id,server.sim.nextSnapshotId)
						//todo:快照已经完成通知模拟器此服务器快照完成
					}else if(len(server.snapshot.messages) < len(server.outboundLinks)){
						server.snapshot.messages = append(server.snapshot.messages,
							&SnapshotMessage{
								src:     src,
								dest:    dest,
								message: MarkerMessage{server.sim.nextSnapshotId}})
					}
				}else{
					log.Printf("%v 到 %v 的信道为空 \n",src,dest)
				}
			}
		}else if(strings.EqualFold(server.snapshotstate,"notstart")){
			//todo:
			// 1 .把这个marker接收存起来
			// 2. 如果快照的状态为 "没开始" 的状态那么就开始快照,对本节点相邻的所有外部信道发送 marker
			server.StartSnapshot(server.sim.nextSnapshotId)
		}else if(strings.EqualFold(server.snapshotstate,"complete")){

		}else{
			log.Fatal("server服务器的状态不能识别")
		}
		server.StartSnapshot(server.sim.nextSnapshotId);
	default:
		log.Fatal("服务器接受到的 message 类型不明确：message = ", message)
	}
}

// Start the chandy-lamport snapshot algorithm on this server.
// This should be called only once per server.
func (server *Server) StartSnapshot(snapshotId int) {
	// TODO: IMPLEMENT ME
	//todo:判断一下server的开始快照的状态
	if(strings.EqualFold(server.snapshotstate,"start")){
		log.Printf(" =============================== %v 服务器 已经开始快照了 ========================== ",server.Id)
	}else {
		//开始快照
		server.snapshotstate = "start"
		server.snapshot.tokens["local"] = server.Tokens //本地快照开始，存储本地的tokens状态
		server.SendToNeighbors(MarkerMessage{snapshotId}) //向相邻的节点发送 marker
		log.Printf("%v 服务器开始了快照",server.Id)
	}
}
