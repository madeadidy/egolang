package main

import (
	"fmt"
	"path/filepath"

	"github.com/disintegration/imaging"
)

func main(){
	src := "tmp/generated.png"
	img, err := imaging.Open(src)
	if err != nil {
		fmt.Println("open error:", err)
		return
	}
	sizes := map[string]int{"xl":1600, "lg":1024, "md":512, "sm":128}
	for k,s := range sizes {
		new := filepath.Join(filepath.Dir(src), fmt.Sprintf("test_%s.png", k))
		d := imaging.Resize(img, s, 0, imaging.Lanczos)
		if err := imaging.Save(d, new); err != nil {
			fmt.Println("save error:", err, new)
		} else {
			fmt.Println("saved", new)
		}
	}
}
