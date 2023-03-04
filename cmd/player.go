package main

import (
	"time"
)

var lastJavascriptoPlay time.Time

func PlayJavascripto() error {
	return nil
	// f, err := os.Open("mata-javascripto.mp3")
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// d, err := mp3.NewDecoder(f)
	// if err != nil {
	// 	return err
	// }

	// c, err := oto.NewContext(d.SampleRate(), 2, 2, 8192)
	// if err != nil {
	// 	return err
	// }
	// defer c.Close()

	// p := c.NewPlayer()
	// defer p.Close()

	// if _, err := io.Copy(p, d); err != nil {
	// 	return err
	// }
	// return nil
}
