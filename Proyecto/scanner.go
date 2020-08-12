package main

import "fmt"

func analizar(entrada string) {
	fmt.Println("Analizando entrada...")
	for _, c := range entrada {
		l := string(c)
		fmt.Println(l)
	}
}
