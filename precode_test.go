package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGeneratorAndWorkers проверяет корректность работы генератора и рабочих горутин.
func TestGeneratorAndWorkers(t *testing.T) {
	chIn := make(chan int64)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var inputSum atomic.Int64   // Сумма сгенерированных чисел
	var inputCount atomic.Int64 // Количество сгенерированных чисел

	go Generator(ctx, chIn, func(i int64) {
		inputSum.Add(i)
		inputCount.Add(1)
	})

	const NumOut = 5
	outs := make([]chan int64, NumOut)
	for i := 0; i < NumOut; i++ {
		outs[i] = make(chan int64)
		go Worker(chIn, outs[i])
	}

	amounts := make([]int64, NumOut)
	chOut := make(chan int64, NumOut)

	var wg sync.WaitGroup

	for i, out := range outs {
		wg.Add(1)
		go func(in <-chan int64, i int64) {
			defer wg.Done()
			for v := range in {
				chOut <- v
				amounts[i]++
			}
		}(out, int64(i))
	}

	go func() {
		wg.Wait()
		close(chOut)
	}()

	var count int64
	var sum int64

	for val := range chOut {
		count++
		sum += val
	}

	// Проверка результатов
	if inputSum.Load() != sum {
		t.Fatalf("Ошибка: суммы чисел не равны: %d != %d\n", inputSum.Load(), sum)
	}
	if inputCount.Load() != count {
		t.Fatalf("Ошибка: количество чисел не равно: %d != %d\n", inputCount.Load(), count)
	}
	for _, v := range amounts {
		inputCount.Add(-v)
	}
	if inputCount.Load() != 0 {
		t.Fatalf("Ошибка: разделение чисел по каналам неверное\n")
	}

	// Вывод результатов в консоль
	t.Logf("Количество чисел: %d %d", inputCount.Load(), count)
	t.Logf("Сумма чисел: %d %d", inputSum.Load(), sum)
	t.Logf("Разбивка по каналам: %v", amounts)
}
