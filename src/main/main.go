package main

import (
	"fmt"
)

func main()  {
	fmt.Println()
}

func test3(){
	snapshots := make([]string, 0,2); //创建了一个切片
	//切片的长度是可以改变的可以追加
	//append(snapshots, "10")
	snapshots = append(snapshots, "111","2","3","4","5")
	mymap := make(map[string]int,100)
	//mymap["n1"] = 1
	src := mymap["n1"]
	fmt.Println(src)
	//fmt.Println(mymap);
	//fmt.Println(snapshots);
	fmt.Println("你好啊")
}

