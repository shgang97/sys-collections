package main

import (
	"fmt"
	"sync"

	"github.com/shgang97/sys-collections/snowflake"
)

var wg sync.WaitGroup

func main() {
	// 初始化
	node, _ := snowflake.Init(1)
	fmt.Println("***************************** start *****************************")
	for i := 0; i < 10; i++ {
		wg.Add(1)
		li := i
		go func() {
			defer wg.Done()
			id, _ := node.NextId()
			fmt.Printf("第%d协程生成的id：%d,二进制为：%b\n", li, id, id)
		}()
	}
	wg.Wait()
	fmt.Println("***************************** end *****************************")
}
