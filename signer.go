package main

import (
	"sort"
	"strconv"
	"sync"
	"time"
)

// сюда писать код

// доделать нужно
func ExecutePipeline() {
	//конвейерную обработку функций-воркеров
	var results []string
	mu := &sync.Mutex{} //нужны чтобы md5 чекать
	ch1 := make(chan string, 1)

	for i := 0; i < 2; i++ {
		go func(i int, mu *sync.Mutex, ch1 chan<- string) {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			data := strconv.Itoa(i)
			ansSingleHash := SingleHash(data, wg, mu)
			wg.Wait()
			ch1 <- MultiHash(ansSingleHash)
		}(i, mu, ch1)
	}
	results = append(results, <-ch1)
	results = append(results, <-ch1)
	answer := CombineResults(results) //вывод конечного результата
}

// готово, вроде как
func SingleHash(data string, wg *sync.WaitGroup, mu *sync.Mutex) string {
	defer wg.Done()

	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	go crc32(ch2, data)

	go func(ch1 chan<- string, data string, mu *sync.Mutex) {
		ch1 <- md5(data, mu)
	}(ch1, data, mu)

	go func(ch1 <-chan string) {
		md5data := <-ch1
		crc32(ch3, md5data)
	}(ch1)

	result1 := <-ch2
	result2 := <-ch3
	result := result1 + "~" + result2
	return result
	//crc32(data)+"~"+crc32(md5(data))

}

//переделать нужно
func MultiHash(data string) string {
	var answer []string
	wg := &sync.WaitGroup{}
	for th := 0; th < 6; th++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, th int, data string, answer []string) {
			defer wg.Done()
			answer = append(answer, crc32(strconv.Itoa(th)+data)) // здесь
		}(wg, th, data, answer)
	}
	time.Sleep(time.Millisecond)
	wg.Wait()
	ans := ""
	for _, value := range answer {
		ans = ans + value
	}
	return ans
}

func CombineResults(data []string) string {
	sort.Strings(data)
	answer := ""
	for _, str := range data {
		answer := answer + "_" + str
	}
	return answer
}

func crc32(ch chan<- string, data string) {
	ch <- DataSignerCrc32(data)
}

func md5(data string, mu *sync.Mutex) string {
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
