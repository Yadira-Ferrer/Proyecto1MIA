package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

//Mkfs : valida el comando
func Mkfs(cmd CommandS) {
	var idp string
	var typef byte = 'a'
	var add int64
	var unit byte = 'k'
	for _, prm := range cmd.Params {
		switch strings.ToLower(prm.Name) {
		case "id":
			idp = delQuotationMark(prm.Value)
		case "type":
			if strings.ToLower(prm.Value) == "fast" {
				typef = 'a'
			} else if strings.ToLower(prm.Value) == "fast" {
				typef = 'e'
			} else {
				fmt.Println("[!~MKFS] Valor invalido para el parametro 'type'...")
			}
		case "add":
			if n, err := strconv.Atoi(prm.Value); err == nil {
				add = int64(n)
			} else {
				fmt.Println("[!~MKFS] El valor del parametro 'add' no es un numero.")
			}
		case "unit":
			if strings.ToLower(prm.Value) == "b" {
				unit = 'b'
			} else if strings.ToLower(prm.Value) == "m" {
				unit = 'm'
			} else {
				unit = 'k'
			}
		}
	}
	fmt.Println("\n===== CREACION DEL SISTEMA DE ARCHIVOS =========================")
	fmt.Println("-id:", idp)
	fmt.Println("-type:", string(typef))
	fmt.Println("-add:", add)
	fmt.Println("-unit:", string(unit))
	if idp != "" {
		MakeFileSystem(idp)
	} else {
		fmt.Println("[!] Hace falta el parametro obligatorio 'id'...")
	}
	fmt.Println("================================================================")
}

//MakeFileSystem : crea el sistema de archivos LWH
func MakeFileSystem(idp string) {
	var flgfound bool = false
	sizep := int64(0)
	//mpartition := Mounted{}
	for _, mp := range sliceMP {
		idm := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
		if idp == idm {
			flgfound = true
			sizep = mp.Part.PartSize
			fmt.Println("Inicio Particion:", mp.Part.PartStart)
			//mpartition = mp
			break
		}
	}
	if flgfound {
		GetNumberOfStructures(sizep)
	} else {
		fmt.Println("[!] La particion", idp, " no se encuentra montada...")
	}
}

//GetNumberOfStructures : obtiene la cantidad de estructuras segun la formula...
func GetNumberOfStructures(sizePartition int64) int64 {
	sizeSB := int64(binary.Size(SuperBoot{}))
	sizeAV := int64(binary.Size(ArbolVirtualDir{}))
	sizeDDir := int64(binary.Size(DetalleDirectorio{}))
	sizeInodo := int64(binary.Size(TablaInodo{}))
	sizeBD := int64(binary.Size(BloqueDeDatos{}))
	sizeLog := int64(binary.Size(Log{}))
	// Referencia
	fmt.Println("Tamaño Particion:", sizePartition)
	fmt.Println("Tamaño SB:", sizeSB)
	fmt.Println("Tamaño AVD:", sizeAV)
	fmt.Println("Tamaño Detalle Dir:", sizeDDir)
	fmt.Println("Tamaño Inodo:", sizeInodo)
	fmt.Println("Tamaño Bloque de Datos:", sizeBD)
	fmt.Println("Tamaño Bitacora (Log):", sizeLog)
	// Calculo del numero de estructuras
	numStructures := int64((sizePartition - (2 * sizeSB)) / (27 + sizeAV + sizeDDir + (5*sizeInodo + (20 * sizeBD) + sizeLog)))
	fmt.Println("Cantidad de estructuras:", numStructures)
	fmt.Println("--------------------------------")

	return 0
}
