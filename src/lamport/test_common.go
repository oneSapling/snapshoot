package lamport

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ==================================
//  Helper methods used in test code
// ==================================

// Directory containing all the test files
const testDir = "test_data"

// Read the topology（拓扑结构） from a ".top" file.
// The expected format of the file is as follows:（文件的预期格式是如下形式的）
// 	- The first line contains number of servers N (e.g. "2")第一行包含了几个sever的数
// 	- The next N lines each contains the server ID and the number of tokens on
// 	  that server, in the form "[serverId] [numTokens]" (e.g. "N1 1")（e.g. 比如）
// 	- The rest of the lines represent unidirectional links in the form "[src dst]" (e.g. "N1 N2")
//  -  剩下的行表示单向传播(e.g. "N1 N2")
//2
//n1 1 [serverId] [numTokens]
//n2 2
//n1 n2 单向传播
//n1 n2

//readTopology 读取拓扑结构（参数为 模拟器 和 .top文件）
func readTopology(fileName string, sim *Simulator) {
	b, err := ioutil.ReadFile(path.Join(testDir, fileName))
	checkError(err)
	lines := strings.FieldsFunc( //FieldsFunc把b按照 \n 分隔开
		string(b),
		func(r rune) bool {
			return r == '\n'
		})

	// Must call this before we start logging
	sim.logger.NewEpoch()

	// Parse topology(拓扑) from lines
	numServersLeft := -1
	for _, line := range lines {
		// Ignore 忽视    comments 评论
		fmt.Println("{ top文件的每一行"+line+" }")
		if strings.HasPrefix(line, "#") {  //判断这一行是否以prefix开头
			continue
		}
		if numServersLeft < 0 {
			numServersLeft, err = strconv.Atoi(line) //字符串转换为数字
			checkError(err)
			continue
		}
		// Otherwise, always expect（期望） 2 tokens
		parts := strings.Fields(line) //以空白字符切分这一行的字符串
		if len(parts) != 2 {
			log.Fatal("Expected 2 tokens in line: ", line)
		}
		if numServersLeft > 0 { // severid 和token数的读取
			// This is a server
			serverId := parts[0]
			numTokens, err := strconv.Atoi(parts[1])
			checkError(err)
			sim.AddServer(serverId, numTokens) //模拟器添加一个服务器
			numServersLeft--
		} else { //遍历最后两行
			// This is a link
			src := parts[0] //遍历发送
			dest := parts[1]
			sim.AddForwardLink(src, dest)
		}
	}
}

// Read the events from a ".events" file and
// inject the events into the simulator. 把事件注入到模拟器中
// The expected(预期) format of the file is as follows:
// - "tick N" indicates（表示） N time steps has elapsed（运行） (default N = 1)
//     标记号 N
// - "send N1 N2 1" indicates that N1 sends 1 token to N2
// - "snapshot N2" indicates
// the beginning of the snapshot process , starting on N2（快照过程的开始，从 N2 开始）
// Note that concurrent（并发） events are indicated by（表示了） the lack of ticks between the events.
// 请注意，并发事件由事件之间缺少点
// This function waits until all the snapshot processes have terminated before returning the snapshots collected.
//  这个函数等待所有的快照进程终止才收集快照信息
//
//send N1 N2 1
//snapshot N2 快照过程的开始，从 N2 开始
//tick default N = 1
//读取 .events文件 的内容并把事件注入到模拟器中

func injectEvents(fileName string, sim *Simulator) []*SnapshotState {
	//path的join函数构造路径
	b, err := ioutil.ReadFile(path.Join(testDir, fileName))
	checkError(err)

	snapshots := make([]*SnapshotState, 0) //初始化一个[]*SnapshotState的切片
	getSnapshots := make(chan *SnapshotState, 100) //创建一个cap为100的*SnapshotState类型的信道
 	numSnapshots := 0 //快照数量

	lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' })
	//按照空格分隔字符串b
	for _, line := range lines {
		// Ignore comments
		if strings.HasPrefix("#", line) { //判断line是否有前缀字符串"#"
			continue
		}

		parts := strings.Fields(line) //按照 空白分隔字符串
		switch parts[0] {
		case "send":
			src := parts[1] //源头
			dest := parts[2] //目的地
			tokens, err := strconv.Atoi(parts[3]) //字符串转换为整数
			checkError(err)
			sim.InjectEvent(PassTokenEvent{src, dest, tokens})
		case "snapshot":
			numSnapshots++
			serverId := parts[1]
			snapshotId := sim.nextSnapshotId
			sim.InjectEvent(SnapshotEvent{serverId})
			go func(id int) { //开启 一个协程
				getSnapshots <- sim.CollectSnapshot(id) //发送SnapshotState 到信道里面
				log.Printf("getSnapshots 从CollectSnapshot函数获得 %v 快照 \n",getSnapshots)
			}(snapshotId)
		case "tick":
			numTicks := 1 //默认tick为1
			if len(parts) > 1 {
				numTicks, err = strconv.Atoi(parts[1])
				checkError(err)
			}
			for i := 0; i < numTicks; i++ {
				sim.Tick()
			}
		default:
			log.Fatal("Unknown event command: ", parts[0])
		}
	}

	// Keep ticking until snapshots complete
	//如果没收到就一直tick
	for numSnapshots > 0 {
		select {
		case snap := <- getSnapshots: //v := <-ch  从Channel ch中接收数据，并将数据赋值给v

			snapshots = append(snapshots, snap)
			numSnapshots--
		default:
			sim.Tick() //tick标记号
		}
	}

	// Keep ticking until we're sure that the last message has been delivered
	//直到收到确认最后一条信息被接收才停止tick
	for i := 0; i < maxDelay+1; i++ {
		sim.Tick()
	}

	return snapshots
}

// Read the state of snapshot from a ".snap" file.(快照状态文件)
// The expected format of the file is as follows:
// 	- The first line contains the snapshot ID (e.g. "0")
// 	- The next N lines contains the server ID and the number of tokens on that server,
// 	  in the form "[serverId] [numTokens]" (e.g. "N1 0"), one line per server
// 	- The rest（余下） of the lines represent messages exchanged between the servers,
// 	  in the form "[src] [dest] [message]" (e.g. "N1 N2 token(1)")
// 读取 .snap 文件
func readSnapshot(fileName string) *SnapshotState {
	b, err := ioutil.ReadFile(path.Join(testDir, fileName)) //读取文件
	checkError(err)
	snapshot := SnapshotState{0, make(map[string]int), make([]*SnapshotMessage, 0)}
	lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' }) //根据空格分隔每一行
	for _, line := range lines {
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 1 {
			// Snapshot ID
			snapshot.id, err = strconv.Atoi(line)
			checkError(err)
		} else if len(parts) == 2 {
			// Server and its tokens
			serverId := parts[0]
			numTokens, err := strconv.Atoi(parts[1])
			checkError(err)
			snapshot.tokens[serverId] = numTokens
		} else if len(parts) == 3 {
			// Src, dest and message
			src := parts[0]
			dest := parts[1]
			messageString := parts[2]
			var message interface{}
			if strings.Contains(messageString, "token") {
				pattern := regexp.MustCompile(`[0-9]+`)
				matches := pattern.FindStringSubmatch(messageString)
				if len(matches) != 1 {
					log.Fatal("Unable to parse token message: ", messageString)
				}
				numTokens, err := strconv.Atoi(matches[0])
				checkError(err)
				message = TokenMessage{numTokens}
			} else {
				log.Fatal("Unknown message: ", messageString)
			}
			snapshot.messages =
				append(snapshot.messages, &SnapshotMessage{src, dest, message})
		}
	}
	return &snapshot
}

// Helper function to pretty print the tokens in the given snapshot state
func tokensString(tokens map[string]int, prefix string) string {
	str := make([]string, 0)
	for _, serverId := range getSortedKeys(tokens) {
		numTokens := tokens[serverId]
		maybeS := "s"
		if numTokens == 1 {
			maybeS = ""
		}
		str = append(str, fmt.Sprintf(
			"%v%v: %v token%v", prefix, serverId, numTokens, maybeS))
	}
	return strings.Join(str, "\n")
}

// Helper function to pretty print the messages in the given snapshot state
func messagesString(messages []*SnapshotMessage, prefix string) string {
	str := make([]string, 0)
	for _, msg := range messages {
		str = append(str, fmt.Sprintf(
			"%v%v -> %v: %v", prefix, msg.src, msg.dest, msg.message))
	}
	return strings.Join(str, "\n")
}

// Assert that the two snapshot states are equal.判断两个快照的状态是否相等如果不相等抛出异常
// If they are not equal, throw an error with a helpful message.
func assertEqual(expected, actual *SnapshotState) {
	if expected.id != actual.id {
		log.Fatalf("Snapshot IDs do not match: %v != %v\n", expected.id, actual.id)
	}
	if len(expected.tokens) != len(actual.tokens) {
		log.Fatalf(
			"Snapshot %v: Number of tokens do not match."+
				"\nExpected:\n%v\nActual:\n%v\n",
			expected.id,
			tokensString(expected.tokens, "\t"),
			tokensString(actual.tokens, "\t"))
	}
	if len(expected.messages) != len(actual.messages) {
		log.Fatalf(
			"Snapshot %v: Number of messages do not match."+
				"\nExpected:\n%v\nActual:\n%v\n",
			expected.id,
			messagesString(expected.messages, "\t"),
			messagesString(actual.messages, "\t"))
	}
	for id, tok := range expected.tokens {
		if actual.tokens[id] != tok {
			log.Fatalf(
				"Snapshot %v: Tokens on %v do not match."+
					"\nExpected:\n%v\nActual:\n%v\n",
				expected.id,
				id,
				tokensString(expected.tokens, "\t"),
				tokensString(actual.tokens, "\t"))
		}
	}
	// Ensure message order is preserved per destination
	// Note that we don't require ordering of messages across all servers to match
	expectedMessages := make(map[string][]*SnapshotMessage)
	actualMessages := make(map[string][]*SnapshotMessage)
	for i := 0; i < len(expected.messages); i++ {
		em := expected.messages[i]
		am := actual.messages[i]
		_, ok1 := expectedMessages[em.dest]
		_, ok2 := actualMessages[am.dest]
		if !ok1 {
			expectedMessages[em.dest] = make([]*SnapshotMessage, 0)
		}
		if !ok2 {
			actualMessages[am.dest] = make([]*SnapshotMessage, 0)
		}
		expectedMessages[em.dest] = append(expectedMessages[em.dest], em)
		actualMessages[am.dest] = append(actualMessages[am.dest], am)
	}
	// Test message order per destination
	for dest := range expectedMessages {
		ems := expectedMessages[dest]
		ams := actualMessages[dest]
		if !reflect.DeepEqual(ems, ams) {
			log.Fatalf(
				"Snapshot %v: Messages received at %v do not match."+
					"\nExpected:\n%v\nActual:\n%v\n",
				expected.id,
				dest,
				messagesString(ems, "\t"),
				messagesString(ams, "\t"))
		}
	}
}

// Helper function to sort the snapshot states by ID.通过id去排序快照的状态
func sortSnapshots(snaps []*SnapshotState) {
	sort.Slice(snaps, func(i, j int) bool {
		s1 := snaps[i]
		s2 := snaps[j]
		return s2.id > s1.id
	})
}

// Verify that the total number of tokens recorded in the snapshot
// preserves（保存） the number of tokens in the system
func checkTokens(sim *Simulator, snapshots []*SnapshotState) {
	expectedTokens := 0
	for _, server := range sim.servers {
		expectedTokens += server.Tokens
	}
	for _, snap := range snapshots {
		snapTokens := 0
		// Add tokens recorded on servers
		for _, tok := range snap.tokens {
			snapTokens += tok
		}
		// Add tokens from messages in-flight
		for _, message := range snap.messages {
			switch msg := message.message.(type) {
			case TokenMessage:
				snapTokens += msg.numTokens
			}
		}
		if expectedTokens != snapTokens {
			log.Fatalf("Snapshot %v: simulator has %v tokens, snapshot has %v:\n%v\n%v",
				snap.id,
				expectedTokens,
				snapTokens,
				tokensString(snap.tokens, "\t"),
				messagesString(snap.messages, "\t"))
		}
	}
}
