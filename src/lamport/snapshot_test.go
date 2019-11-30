package lamport

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func runTest(t *testing.T, topFile string, eventsFile string, snapFiles []string) {
	//startMessage := fmt.Sprintf("{测试用的文件是 《'%v'》, 《'%v'》}", topFile, eventsFile)
	log.Printf("{测试用的文件是 《'%v'》, 《'%v'》}", topFile, eventsFile)
	if debug {
		bars := "=================================================================="
		//startMessage := fmt.Sprintf("%v\n%v\n%v\n", bars, startMessage, bars)
		log.Printf("%v\n", bars)
	}

	// Initialize simulator (初始化模拟器)
	rand.Seed(8053172852482175524) //Seed使用提供的seed值将发生器初始化为确定性状态
	sim := NewSimulator() //创建一个模拟器,分发的工作都是由模拟器来做的
	readTopology(topFile, sim) //读取节点的原始数据，并把原始数据放到模拟器中
	actualSnaps := injectEvents(eventsFile, sim) //读取事件数据
	if len(actualSnaps) != len(snapFiles) {
		t.Fatalf("预期有 %v 个snapshot(s), 得到了got %v\n", len(snapFiles), len(actualSnaps))
	}
	// Optionally print events for debugging 可选择的打印调试事件
	if debug {
		sim.logger.PrettyPrint()
		fmt.Println()
	}
	// Verify that the number of tokens are preserved in the snapshots
	//确认快照中保留了令牌的数量
	checkTokens(sim, actualSnaps)
	// Verify against golden files
	// 和golden files核对一下
	expectedSnaps := make([]*SnapshotState, 0)
	//创建一个*SnapshotState的切片 ，SnapshotState的作用就是记录快照过程的状态
	for _, snapFile := range snapFiles {
		expectedSnaps = append(expectedSnaps, readSnapshot(snapFile))
	}

	sortSnapshots(actualSnaps)
	sortSnapshots(expectedSnaps)
	for i := 0; i < len(actualSnaps); i++ {
		assertEqual(expectedSnaps[i], actualSnaps[i])
	}
}
//t *testing.T是一个测试类 go test -run + 类名可以运行
func Test2Nodes(t *testing.T) {
	runTest(t, "2nodes.top", "2nodes-simple.events", []string{"2nodes-simple.snap"})
}

func Test2NodesSingleMessage(t *testing.T) {
	runTest(t, "2nodes.top", "2nodes-message.events", []string{"2nodes-message.snap"})
}

func Test3NodesMultipleMessages(t *testing.T) {
	runTest(t, "3nodes.top", "3nodes-simple.events", []string{"3nodes-simple.snap"})
}

func Test3NodesMultipleBidirectionalMessages(t *testing.T) {
	runTest(
		t,
		"3nodes.top",
		"3nodes-bidirectional-messages.events",
		[]string{"3nodes-bidirectional-messages.snap"})
}

func Test8NodesSequentialSnapshots(t *testing.T) {
	runTest(
		t,
		"8nodes.top",
		"8nodes-sequential-snapshots.events",
		[]string{
			"8nodes-sequential-snapshots0.snap",
			"8nodes-sequential-snapshots1.snap",
		})
}

func Test8NodesConcurrentSnapshots(t *testing.T) {
	runTest(
		t,
		"8nodes.top",
		"8nodes-concurrent-snapshots.events",
		[]string{
			"8nodes-concurrent-snapshots0.snap",
			"8nodes-concurrent-snapshots1.snap",
			"8nodes-concurrent-snapshots2.snap",
			"8nodes-concurrent-snapshots3.snap",
			"8nodes-concurrent-snapshots4.snap",
		})
}

func Test10NodesDirectedEdges(t *testing.T) {
	runTest(
		t,
		"10nodes.top",
		"10nodes.events",
		[]string{
			"10nodes0.snap",
			"10nodes1.snap",
			"10nodes2.snap",
			"10nodes3.snap",
			"10nodes4.snap",
			"10nodes5.snap",
			"10nodes6.snap",
			"10nodes7.snap",
			"10nodes8.snap",
			"10nodes9.snap",
		})
}
//-------------------------------------------------------
func gotest(i string,string string){
	fmt.Println("{"+i+"},"+string);
}
/*
	//arr := make([]*string,0)
	//a := "10"
	//append(arr,&a)
	//log.Print(arr,a)
*/

func TestZs(t *testing.T){
	done := make(chan bool)
	go hello(done)
	<-done //在收到done的之前这个程序一直是阻塞的
	fmt.Println("main function")
}

func hello(done chan bool) {
	fmt.Println("Hello world goroutine")
	done <- true
}
func Test休眠函数(t *testing.T){
	done := make(chan bool)
	fmt.Println("Main going to call hello go goroutine")
	go hello2(done)
	<-done
	fmt.Println("Main received data")
}
func hello2(done chan bool) {
	fmt.Println("hello go routine is going to sleep")
	time.Sleep(10 * time.Second)
	fmt.Println("hello go routine awake and going to write to done")
	done <- true
}

func Test协程1(t *testing.T){
	for i:=0; i<10; i++ {
		go Add(i, i)
	}
}
func Add(x, y int) {
	z := x + y
	fmt.Println(z)
}
func Test协程2(t *testing.T){
	lock := &sync.Mutex{}
	for i:=0; i<10; i++ {
		go Count(lock)
	}
	for {
		lock.Lock()
		c := counter
		lock.Unlock()
		runtime.Gosched()
		if c >= 10 {
			break
		}
	}
}
var counter int = 0
func Count(lock *sync.Mutex) {
	lock.Lock()
	counter++
	fmt.Println("counter =", counter)
	lock.Unlock()
}

func Test携程3(t *testing.T){
	b := 2
	channal := make(chan string,10)
	channal2 := make(chan string,20)
	go func(a int) {
		channal <- hello5(channal,channal2)
	}(b)
	log.Print(<-channal)
	log.Print("main线程")
}
func hello5(a chan string,b chan string) string{
	//log.Print("子协程")
	//time.Sleep(10 * time.Second)
	select {
	case <- a:
		log.Print("从chan1里面读取到数据")
		return "从chan1里面读取到数据"
	case <- b:
		log.Print("从chan2里面读取到数据")
		return "从chan2里面读取到数据"
	default:
		log.Print("没有读取到")
		return "没有读取到"
	}
	return "5"

}