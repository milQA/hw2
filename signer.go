package main

import (
	"runtime"
	"sort"
	"strconv"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	ch := make([]chan interface{}, len(jobs)+1)
	for i := 0; i < len(jobs)+1; i++ {
		ch[i] = make(chan interface{}, 6)
	}
	for i, jb := range jobs {
		in := ch[i]
		out := ch[i+1]
		wg.Add(1)
		go func(j job, in, out chan interface{}, wg *sync.WaitGroup) {
			j(in, out)
			close(out)
			wg.Done()
		}(jb, in, out, wg)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	mu2 := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for input := range in {
		mu2.Lock()
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			ch2 := make(chan string)
			ch3 := make(chan string)
			datain := input.(int)
			data := strconv.Itoa(datain)
			mu.Lock()
			md5data := DataSignerMd5(data)
			mu.Unlock()

			go fcrc32(ch2, md5data)
			go fcrc32(ch3, data)

			result1 := <-ch3
			result2 := <-ch2
			result := result1 + "~" + result2
			out <- result
			wg.Done()
		}(wg)
		mu2.Unlock()
		runtime.Gosched()
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			data := input.(string)
			var answer []string
			ch := make([]chan string, 6)
			for th := 0; th < 6; th++ {
				ch[th] = make(chan string)
				go fcrc32(ch[th], strconv.Itoa(th)+data)
			}
			for th := 0; th < 6; th++ {
				answer = append(answer, <-ch[th])
			}
			ans := ""
			for _, value := range answer {
				ans = ans + value
			}
			out <- ans
			wg.Done()
		}(wg)
		runtime.Gosched()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var data []string
	for input := range in {
		dataOne := input
		data = append(data, dataOne.(string))
	}
	sort.Strings(data)
	answer := data[0]
	for i := 1; i < len(data); i++ {
		answer = answer + "_" + data[i]
	}
	out <- answer
}

func fcrc32(ch chan<- string, data string) {
	temp := DataSignerCrc32(data)
	ch <- temp
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
