package main

import ( 
	"fmt"
	"log" 
	"net/http" 
	"sync"
	"image"
	"image/color" 
	"image/gif" 
	"io" 
	"math" 
	"math/rand" 
	"time"
	"strconv"
)

var palette = []color.Color{color.Black, color.RGBA{0x00,0xff,0x00,0xff}}

const ( 
	blacklndex = 0 
	greenlndex = 1 
)

var mu sync.Mutex 
var count int

func main() { 
	http.HandleFunc("/", handler) 
	http.HandleFunc("/count", counter)
	http.HandleFunc("/lissajous", liss) 
	log.Fatal(http.ListenAndServe("localhost:8000", nil)) 
}

func handler(w http.ResponseWriter, r *http.Request) { 
	mu.Lock()
	count++ 
	mu.Unlock() 
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto) 
	for k, v := range r.Header { 
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v) 
	} 
	fmt.Fprintf(w, "Host = %q\n", r.Host) 
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr) 
	if err := r.ParseForm(); err != nil { 
		log.Print(err) 
	} 
	for k, v := range r.Form { 
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v) 
	}
}

func counter(w http.ResponseWriter, r *http.Request) { 
	mu.Lock() 
	fmt.Fprintf(w, "Count %d\n", count) 
	mu.Unlock() 
}

func liss(w http.ResponseWriter, r *http.Request) { 
	cycles := 5
	if err := r.ParseForm(); err != nil { 
		log.Print(err) 
	}
	if val, ok := r.Form["cycles"]; ok {
	    if tmpCycles, atoiErr := strconv.Atoi(val[0]); atoiErr == nil {
	    	cycles = tmpCycles
	    }
	} 
	lissajous(w, cycles)
}

func lissajous(out io.Writer, cycles int) {
	const ( 
		res = 0.001 // Угловое разрешение 
		size = 100 // Канва изображения охватывает [size..+size] 
		nframes = 64 // Количество кадров анимации 
		delay = 8 // Задержка между кадрами (единица - 10мс) 
	) 

	rand.Seed(time.Now().UTC().UnixNano()) 
	freq := rand.Float64() * 3.0 // Относительная частота колебаний у 
	anim := gif.GIF{LoopCount: nframes}
	phase := 0.0 // Разность фаз 

	for i := 0; i < nframes; i++ { 
		rect := image.Rect(0, 0, 2*size+1, 2*size+1) 
		img := image.NewPaletted(rect, palette) 
		for t := 0.0; t < float64(cycles*2)*math.Pi; t += res {
			x := math.Sin(t) 
			y := math.Sin(t*freq + phase) 
			img.SetColorIndex(size+int(x*size+0.5), size+int(y*size+0.5), greenlndex) 
		} 
		phase += 0.1 
		anim.Delay = append(anim.Delay, delay) 
		anim.Image = append(anim.Image, img) 
	} 

	gif.EncodeAll(out, &anim) 
}