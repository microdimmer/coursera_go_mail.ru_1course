package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ExecutePipeline ...
func ExecutePipeline(freeFlowJobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 0) //для первой горутины где нет вхожных данных
	for _, valJob := range freeFlowJobs {
		wg.Add(1)
		out := make(chan interface{}, 100) //Вы можете ожидать, что у вас никогда не будет более 100 элементов во входных данных
		go startWorker(wg, valJob, in, out)
		in = out
	}
	wg.Wait()
}

// startWorker ...
func startWorker(wg *sync.WaitGroup, freeFlowJob job, in chan interface{}, out chan interface{}) {
	defer wg.Done()
	freeFlowJob(in, out)
	close(out)
}

type dataStruct struct {
	indexNum   int
	dataString string
}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~), где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	outChannelCRC32MD5 := make(chan dataStruct, 100) // data2 := DataSignerCrc32(md5Sum)
	outChannelCRC32 := make(chan dataStruct, 100)    // data1 := DataSignerCrc32(data)

	mut := &sync.Mutex{}
	wg1 := &sync.WaitGroup{}
	var i int
	for idata := range in {
		data := fmt.Sprintf("%v", idata) //data := (<-in).(string)
		// fmt.Println(i, "SingleHash data", data)
		outChannelMD5 := make(chan string, 100) // md5Sum := DataSignerMd5(data)

		go func(jdata string, jout chan string) { //считает MD5()
			mut.Lock()
			md5Sum := DataSignerMd5(jdata)
			mut.Unlock()
			// fmt.Println(i, "SingleHash md5(data)", md5Sum) //0 SingleHash md5(data) cfcd208495d565ef66e7dff9f98764da
			jout <- md5Sum
			// fmt.Println("md5Sum routine finish")
		}(data, outChannelMD5)

		wg1.Add(1)
		go func(jwg1 *sync.WaitGroup, jin chan string, jout chan dataStruct, ji int) { //считает CRC32(MD5)
			defer jwg1.Done()
			jdata := <-jin
			crc32MD5 := DataSignerCrc32(jdata)
			jout <- dataStruct{indexNum: ji, dataString: crc32MD5}
			// jout <- crc32MD5
		}(wg1, outChannelMD5, outChannelCRC32MD5, i)

		go func(jdata string, jout chan dataStruct, ji int) { //считает CRC32()
			crc32Sum := DataSignerCrc32(jdata)
			jout <- dataStruct{indexNum: ji, dataString: crc32Sum}
		}(data, outChannelCRC32, i)

		i++
	}
	wg1.Wait() //ждём пока закончатся все вычисления CRC32 и MD5 - crc32(data) и crc32(md5(data))
	close(outChannelCRC32MD5)
	close(outChannelCRC32)

	crc32md5dataMap := make(map[int]string, 100)
	crc32DataMap := make(map[int]string, 100)

	for chData := range outChannelCRC32MD5 {
		crc32md5dataMap[chData.indexNum] = chData.dataString
	}
	for chData := range outChannelCRC32 {
		crc32DataMap[chData.indexNum] = chData.dataString
	}
	for key := range crc32md5dataMap {
		outData := crc32DataMap[key] + "~" + crc32md5dataMap[key]
		out <- outData
	}

	// md5Sum := DataSignerMd5(data)
	// data2 := DataSignerCrc32(md5Sum)
	// data1 := DataSignerCrc32(data)
	// out <- data1 + "~" + data2

}

type dataStruct2 struct {
	singleHashIndex int
	dataS           dataStruct
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), где th=0..5 ( т.е. 6 хешей на каждое входящее значение ),
// потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) { //4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
	wg := &sync.WaitGroup{}
	outChannelCRC32 := make(chan dataStruct2, 100) // md5Sum := DataSignerMd5(data)
	var ii int
	for idata := range in { //обрабатывает каждое значение приходящее в канал
		data := fmt.Sprintf("%v", idata) //переводим в string
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(jwg *sync.WaitGroup, jout chan dataStruct2, ji int, jii int) {
				defer jwg.Done()
				jdata := DataSignerCrc32(fmt.Sprintf("%v", ji) + data)
				fmt.Println(data, "MultiHash: crc32(th+step1))", ji, jdata)
				jout <- dataStruct2{singleHashIndex: jii, dataS: dataStruct{indexNum: ji, dataString: jdata}} //возвращаем i потока для сортировки и вычисленные данные
			}(wg, outChannelCRC32, i, ii)
		}
		ii++
	}
	wg.Wait() // ждём высчитывания всех CRC32
	close(outChannelCRC32)

	var dataSl []dataStruct2
	for outI := range outChannelCRC32 { //получаем данные из канала
		dataSl = append(dataSl, outI)
	}
	sort.SliceStable(dataSl, func(i, j int) bool { //сортируем выходные данные
		if dataSl[i].singleHashIndex < dataSl[j].singleHashIndex {
			return true
		}
		if dataSl[i].singleHashIndex > dataSl[j].singleHashIndex {
			return false
		}
		return (dataSl[i].dataS.indexNum < dataSl[j].dataS.indexNum)
	})
	// fmt.Printf("\ndataSl %v\n", dataSl)

	var outString string
	var iiiiiii int
	for _, value := range dataSl {
		iiiiiii++
		outString = outString + value.dataS.dataString
		if iiiiiii == 6 {
			out <- outString
			outString = ""
			iiiiiii = 0
		}
	}
}

// CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/), объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	var intSl []string
	for i := range in { //получаем данные из канала
		intSl = append(intSl, fmt.Sprintf("%v", i))
	}
	sort.Strings(intSl)
	out <- strings.Join(intSl, "_")
}
