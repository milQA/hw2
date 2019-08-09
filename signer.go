package main

import (
	"sort"
	"strconv"
	"sync"
	"time"
	"fmt"
	"sync/atomic"
)

// сюда писать код

// доделано
func ExecutePipeline(in []string) {
	//конвейерную обработку функций-воркеров
	var results []string
	mu := &sync.Mutex{} //нужны чтобы md5 чекать
	ch1 := make(chan string, 1)
	inCount := len(in)
	for _, i := range in {
		go func(i string, mu *sync.Mutex, ch1 chan<- string) {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			//data := strconv.Itoa(i)
			ansSingleHash := SingleHash(i /*data,*/, wg, mu)
			wg.Wait()
			ch1 <- MultiHash(ansSingleHash)
		}(i, mu, ch1)
	}
	for i := 0; i < inCount; i++ {
		results = append(results, <-ch1)
	}
	answer := CombineResults(results) //вывод конечного результата
	//time.Sleep(5*time.Second)
	fmt.Println("=====================")
	fmt.Println("EP:")
	fmt.Println(answer)
}

// готово
func SingleHash(data string, wg *sync.WaitGroup, mu *sync.Mutex) string {
	defer wg.Done()

	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	go fcrc32(ch2, data)

	go func(ch1 chan<- string, data string, mu *sync.Mutex) {
		ch1 <- fmd5(data, mu)
	}(ch1, data, mu)

	go func(ch3 chan<- string, ch1 <-chan string) {
		md5data := <-ch1
		fcrc32(ch3, md5data)
	}(ch3, ch1)

	result1 := <-ch2
	result2 := <-ch3
	result := result1 + "~" + result2
	fmt.Println(result) //norm. без перегревов
	return result
	//crc32(data)+"~"+crc32(md5(data))

}

//сделано. Выводит MH result
func MultiHash(data string) string {
	var answer []string
	ch1 := make(chan string)
	//wg := &sync.WaitGroup{}
	for th := 0; th < 6; th++ {
		//wg.Add(1)
		go func( /*wg *sync.WaitGroup,*/ th int, data string) {
			//defer wg.Done()
			fcrc32(ch1, strconv.Itoa(th)+data) // здесь
			
		}( /*wg,*/ th, data)
		time.Sleep(time.Millisecond)
	}
	for th := 0; th < 6; th++ {
		answer = append(answer, <-ch1)
		fmt.Println(th, "||", answer[th])
	}
	//time.Sleep(time.Millisecond)
	//wg.Wait() //ожидание, что все MH отработают
	ans := ""
	for _, value := range answer {
		ans = ans + value
	}
	//fmt.Println(ans)
	return ans
}

func CombineResults(data []string) string {
	fmt.Println("=====")
	fmt.Println(data)
	fmt.Println("=====")
	sort.Strings(data)
	answer := data[0]
	for i := 1; i < len(data); i++ {
		answer = answer + "_" + data[i]
	}
	return answer
}

func fcrc32(ch chan<- string, data string) {
	temp := DataSignerCrc32(data)
	//fmt.Println("crc32 ",temp)
	ch <- temp
}

func fmd5(data string, mu *sync.Mutex) string {
	mu.Lock()
	defer mu.Unlock()
	return DataSignerMd5(data)
}
/*
0 SingleHash data 0
0 SingleHash md5(data) cfcd208495d565ef66e7dff9f98764da
0 SingleHash crc32(md5(data)) 502633748
0 SingleHash crc32(data) 4108050209
0 SingleHash result 4108050209~502633748
4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
4108050209~502633748 MultiHash: crc32(th+step1)) 1 803518384
4108050209~502633748 MultiHash: crc32(th+step1)) 2 1425683795
4108050209~502633748 MultiHash: crc32(th+step1)) 3 3407918797
4108050209~502633748 MultiHash: crc32(th+step1)) 4 2730963093
4108050209~502633748 MultiHash: crc32(th+step1)) 5 1025356555
4108050209~502633748 MultiHash result: 29568666068035183841425683795340791879727309630931025356555
1 SingleHash data 1
1 SingleHash md5(data) c4ca4238a0b923820dcc509a6f75849b
1 SingleHash crc32(md5(data)) 709660146
1 SingleHash crc32(data) 2212294583
1 SingleHash result 2212294583~709660146
2212294583~709660146 MultiHash: crc32(th+step1)) 0 495804419
2212294583~709660146 MultiHash: crc32(th+step1)) 1 2186797981
2212294583~709660146 MultiHash: crc32(th+step1)) 2 4182335870
2212294583~709660146 MultiHash: crc32(th+step1)) 3 1720967904
2212294583~709660146 MultiHash: crc32(th+step1)) 4 259286200
2212294583~709660146 MultiHash: crc32(th+step1)) 5 2427381542
2212294583~709660146 MultiHash result: 4958044192186797981418233587017209679042592862002427381542
CombineResults 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
*/
