package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// SizeStructures : almacena el tamaño de cada estructura del sistema LWH
type SizeStructures struct {
	sizeSB    int64
	sizeAV    int64
	sizeDDir  int64
	sizeInodo int64
	sizeBD    int64
	sizeLog   int64
}

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
		MakeFileSystem(idp, typef)
	} else {
		fmt.Println("[!] Hace falta el parametro obligatorio 'id'...")
	}
	fmt.Println("================================================================")
}

//MakeFileSystem : crea el sistema de archivos LWH
func MakeFileSystem(idp string, typef byte) {
	var flgfound bool = false
	mpartition := Mounted{}
	for _, mp := range sliceMP {
		idm := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
		if idp == idm {
			flgfound = true
			mpartition = mp
			break
		}
	}
	if flgfound {
		partition := mpartition.Part
		// Se realiza el formateo de la partición
		if typef == 'u' {
			writeByteArray(mpartition.Path, partition.PartStart, partition.PartSize)
		}
		// Almacena la posicion actual
		var cpd int64 // Current Position Disk Partition
		// Obtener el nombre del disco ¿?
		// Se obtiene el tamaño de las estructuras y la cantidad (#Estructuras)
		sStrc, cStrc := GetNumberOfStructures(partition.PartSize)
		// Se creará el Super Boot
		newSB := SuperBoot{}
		copy(newSB.nombreHd[:], mpartition.Name)
		newSB.arbolesVirtualesLibres = cStrc   // AVD-free = #Estructuras
		newSB.detallesDirectorioLibres = cStrc // DDir-free = #Estructuras
		newSB.inodosLibres = cStrc             // Inodos-free = #Estructuras
		newSB.bloquesLibres = cStrc            // Bloques-free = #Estructuras
		newSB.fechaCreacion = getCurrentTime()
		newSB.fechaUltimoMontaje = mpartition.TMount
		newSB.conteoMontajes = 1
		// Inicio BMap AVD = Inicio_Particion + SizeSB
		cpd = partition.PartStart + sStrc.sizeSB
		newSB.aptBmapArbolDirectorio = cpd
		// Inicio AVD = Inicio BitMap AVD + #Estructuras
		cpd = cpd + cStrc
		newSB.aptArbolDirectorio = cpd
		// Inicio BMap DDir = Inicio AVD + (sizeAVD*#Estructuras)
		cpd = cpd + (sStrc.sizeAV * cStrc)
		newSB.aptBmapDetalleDirectorio = cpd
		// Inicio DDir = Inicio BMap DDir + #Estructuras
		cpd = cpd + cStrc
		newSB.aptDetalleDirectorio = cpd
		// Inicio BMap Inodo = Inicio DDir + (sizeDDir * #Estructuras)
		cpd = cpd + (sStrc.sizeDDir * cStrc)
		newSB.aptBmapTablaInodo = cpd
		// Inicio Inodos = Inicio BMap Inodo + (5 * sizeInodo)
		cpd = cpd + (5 * sStrc.sizeInodo)
		newSB.aptTablaInodo = cpd
		// Inicio BMap Bloque = Inicio Inodos + (5 * sizeInodo * #Estructuras)
		cpd = cpd + (5 * sStrc.sizeInodo * cStrc)
		newSB.aptBmapBloques = cpd
		// Inicio Bloque = Inicio Inodo + (20 * #Estructuras)
		cpd = cpd + (20 * cStrc)
		newSB.aptBloques = cpd
		// Inicio Bitacora (Log) = Inicio Bloque + (20 * sizeBloque * #Estructuras)
		cpd = cpd + (20 * sStrc.sizeBD * cStrc)
		newSB.aptLog = cpd
		// Inicio Copia SB = Inicio Bitacora + (sizeLog * #Estructuras)
		cpd = cpd + (sStrc.sizeLog * cStrc)
		newSB.aptLog = cpd
		//--- Se guarda el tamaño de las estructuras ------------------------------------
		newSB.tamStrcArbolDirectorio = sStrc.sizeAV
		newSB.tamStrcDetalleDirectorio = sStrc.sizeDDir
		newSB.tamStrcInodo = sStrc.sizeInodo
		newSB.tamStrcBloque = sStrc.sizeBD
		//--- Se guarda el primer bit vacio del bitmap de cada estructura ---------------
		newSB.primerBitLibreArbolDir = newSB.aptBmapArbolDirectorio
		newSB.primerBitLibreDetalleDir = newSB.aptBmapDetalleDirectorio
		newSB.primerBitLibreTablaInodo = newSB.aptBmapTablaInodo
		newSB.primerBitLibreBloques = newSB.aptBmapBloques
		//--- Numero Magico -------------------------------------------------------------
		newSB.numeroMagico = 201503442
		//--- Escribir SB en Disco ------------------------------------------------------
		WriteSuperBoot(mpartition.Path, newSB, partition.PartStart, sStrc.sizeSB)
	} else {
		fmt.Println("[!] La particion", idp, " no se encuentra montada...")
	}
}

//GetNumberOfStructures : obtiene la cantidad de estructuras segun la formula...
func GetNumberOfStructures(sizePartition int64) (SizeStructures, int64) {
	sos := SizeStructures{}
	sos.sizeSB = int64(binary.Size(SuperBoot{}))
	sos.sizeAV = int64(binary.Size(ArbolVirtualDir{}))
	sos.sizeDDir = int64(binary.Size(DetalleDirectorio{}))
	sos.sizeInodo = int64(binary.Size(TablaInodo{}))
	sos.sizeBD = int64(binary.Size(BloqueDeDatos{}))
	sos.sizeLog = int64(binary.Size(Log{}))
	// Referencia
	/* fmt.Println("Tamaño Particion:", sizePartition)
	fmt.Println("Tamaño SB:", sizeSB)
	fmt.Println("Tamaño AVD:", sizeAV)
	fmt.Println("Tamaño Detalle Dir:", sizeDDir)
	fmt.Println("Tamaño Inodo:", sizeInodo)
	fmt.Println("Tamaño Bloque de Datos:", sizeBD)
	fmt.Println("Tamaño Bitacora (Log):", sizeLog) */
	// Calculo del numero de estructuras
	numStructures := int64((sizePartition - (2 * sos.sizeSB)) / (27 + sos.sizeAV + sos.sizeDDir + (5*sos.sizeInodo + (20 * sos.sizeBD) + sos.sizeLog)))
	//fmt.Println("Cantidad de estructuras:", numStructures)
	return sos, numStructures
}

// WriteSuperBoot : escribe el super boot en la particion
func WriteSuperBoot(path string, sboot SuperBoot, position int64, size int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	sb := &sboot
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, sb)
	state := writeBytes(file, bin.Bytes())
	if state {
		fmt.Println("*** Se ha escrito SuperBoot  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el SuperBoot.")
	}
}

// ReadSuperBoot : recupera el superboot de la particion
func ReadSuperBoot(path string) SuperBoot {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Println(err)
	}
	// Se declara MBR contenedor
	recSb := SuperBoot{}
	// Se obtiene el tamaño del MBR
	var sizeSb int64 = int64(binary.Size(recSb))
	// Lectura los bytes determinados por mbrSize
	data := readBytes(file, sizeSb)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &recSb)
	if err != nil {
		log.Fatal("[!] Fallo la lectura del Super Boot", err)
	}
	return recSb
}
