package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"unicode"
)

// dot -Tpng filename.dot -o outfile.png

//MbrReport : genera el script del reporte
func MbrReport(path string, mbr MBR) {
	p, n, e := SeparatePath(path)
	t := mbr.MbrTime
	fyh := strconv.FormatInt(t.Day, 10) + "/" + strconv.FormatInt(t.Month, 10) + "/" + strconv.FormatInt(t.Year, 10) + " " + strconv.FormatInt(t.Hour, 10) + ":" + strconv.FormatInt(t.Minute, 10) + ":" + strconv.FormatInt(t.Seconds, 10)
	cdot := "digraph G {\ntbl [ shape=plaintext label=<\n<table border='0' color='grey' cellspacing='0' cellpadding=\"5\">\n<tr><td colspan=\"2\" bgcolor=\"#1b1f40\"><font color=\"white\">MBR</font></td></tr>\n"
	cdot += "<tr><td width=\"120\" align=\"left\">MBR Tamaño:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(mbr.MbrSize, 10) + "</td></tr>\n"
	cdot += "<tr><td align=\"left\">MBR Fecha:</td><td align=\"left\">" + fyh + "</td></tr>\n"
	cdot += "<tr><td align=\"left\">MBR Firma:</td><td align=\"left\">" + strconv.FormatInt(mbr.MbrDiskSignature, 10) + "</td></tr>\n"
	cdot += "<tr><td align=\"left\">MBR Ajuste:</td><td align=\"left\">Primer</td></tr>\n"
	// A partir de acá se genera la información de las particiones
	for i, p := range mbr.MbrPartitions {
		cdot += "<tr><td colspan=\"2\" align=\"left\" bgcolor=\"#81BEF7\">Particion " + strconv.Itoa(i+1) + "</td></tr>\n"
		cdot += "<tr><td align=\"left\">Nombre: </td><td align=\"left\">" + GetString(p.PartName) + "</td></tr>\n"
		if p.PartStatus == 1 {
			cdot += "<tr><td align=\"left\">Estado: </td><td align=\"left\">1</td></tr>\n"
		} else {
			cdot += "<tr><td align=\"left\">Estado: </td><td align=\"left\">0</td></tr>\n"
		}
		cdot += "<tr><td align=\"left\">Tipo: </td><td align=\"left\">" + GetType(p.PartType) + "</td></tr>\n"
		cdot += "<tr><td align=\"left\">Ajuste: </td><td align=\"left\">" + GetFit(p.PartFit) + "</td></tr>\n"
		cdot += "<tr><td align=\"left\">Tamaño: </td><td align=\"left\">" + strconv.FormatInt(p.PartSize, 10) + "</td></tr>\n"
		cdot += "<tr><td align=\"left\">Inicio: </td><td align=\"left\">" + strconv.FormatInt(p.PartStart, 10) + "</td></tr>\n"
	}
	cdot += "</table>>];}"
	// Se escribirá el archivo dot
	state := WriteFile(p, n, "dot", cdot)
	if state {
		fmt.Println("> DOT escrito exitosamente...")
		outType := "-T" + e
		pathDot := p + n + ".dot"
		pathRep := path
		DotGenerator(outType, pathDot, pathRep)
	}
}

// DiskReport : genera el reporte del disco
func DiskReport(path string, mbr MBR, dskpath string) {
	p, n, e := SeparatePath(path)
	cdot := "digraph test {\ngraph [ratio=fill];\nnode [label=\"\\N\", fontsize=15, shape=plaintext];\ngraph [bb=\"0,0,352,154\"];\narset [label=<\n<TABLE ALIGN=\"CENTER\">\n<TR>\n<TD>MBR</TD>\n"
	for _, p := range mbr.MbrPartitions {
		if p.PartStatus == 0 {
			cdot += "<TD>\n<TABLE BORDER=\"0\">\n<TR><TD>Libre</TD></TR>\n</TABLE>\n</TD>\n"
		} else {
			if p.PartType == 'p' {
				cdot += "<TD>\n<TABLE BORDER=\"0\">\n<TR><TD>Primaria<br/>" + GetString(p.PartName) + "</TD></TR>\n</TABLE>\n</TD>\n"
			} else if p.PartType == 'e' {
				cdot += "<TD>\n<TABLE BORDER=\"0\">\n<TR><TD>Extendida<br/>" + GetString(p.PartName) + "</TD></TR>\n<TR>\n<TD>\n<TABLE BORDER=\"1\">\n<TR>\n"
				position := p.PartStart
				for true {
					ebr := readEBR(dskpath, position)
					if ebr.PartStatus == 0 {
						break
					} else {
						cdot += "<TD>EBR</TD>\n<TD>\n<TABLE BORDER=\"0\">\n<TR><TD>Logica<br/>" + GetString(ebr.PartName) + "</TD></TR>\n</TABLE>\n</TD>\n"
						position = ebr.PartNext
					}
				}
				cdot += "</TR>\n</TABLE>\n</TD>\n</TR>\n</TABLE>\n</TD>\n"
			}
		}
	}
	cdot += "</TR>\n</TABLE>\n>, ];\n}"
	state := WriteFile(p, n, "dot", cdot)
	if state {
		fmt.Println("> DOT escrito exitosamente...")
		outType := "-T" + e
		pathDot := p + n + ".dot"
		pathRep := path
		DotGenerator(outType, pathDot, pathRep)
	}
}

// SuperBootReport : genera el reporte del Super Boot
func SuperBootReport(path string, mp Mounted) {
	p, n, e := SeparatePath(path)
	superBoot := ReadSuperBoot(mp.Path, mp.Part.PartStart)
	cDot := "digraph G {\ntbl [ shape=plaintext label=<\n<table border='1' color='grey' cellspacing='0' cellpadding=\"5\">\n<tr>\n<td width=\"120\" bgcolor=\"#1b1f40\"><font color=\"white\">NOMBRE</font></td>\n<td width=\"130\" bgcolor=\"#1b1f40\"><font color=\"white\">VALOR</font></td>\n</tr>\n"
	// Escritura de los atributos de SB
	cDot += "<tr><td width=\"120\" align=\"left\">Nombre HD:</td><td width=\"130\" align=\"left\">" + GetString(superBoot.NombreHd) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad AV:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.CantArbolVirtual, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Detalle Directorio:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.CantDetalleDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Inodos:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.CantidadInodos, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Bloques:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.CantidadBloques, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad AV Libres:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.ArbolesVirtualesLibres, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Detalle Dir Libres:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.DetallesDirectorioLibres, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Inodos Libres:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.InodosLibres, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad Bloques Libres:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.BloquesLibres, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Fecha Creación:</td><td width=\"130\" align=\"left\">" + GetDateAsString(superBoot.FechaCreacion) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Fecha Ultimo Montaje:</td><td width=\"130\" align=\"left\">" + GetDateAsString(superBoot.FechaUltimoMontaje) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Cantidad de montajes:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.ConteoMontajes, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt BitMap AVD:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptBmapArbolDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt AVD:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptArbolDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt BitMap Detalle Dir:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptBmapDetalleDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt Detalle Directorio:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptDetalleDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt BitMap Tabla Inodo:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptBmapTablaInodo, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt Tabla Inodo:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptTablaInodo, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt BitMap Bloques:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptBmapBloques, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt Bloques de Datos:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptBloques, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Apt Bitacora (Log):</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.AptLog, 10) + "</td></tr>\n"
	// Tamaños de las estructuras
	cDot += "<tr><td width=\"120\" align=\"left\">Tamaño Struct AVD:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.TamStrcArbolDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Tamaño Struct Detalle Dir:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.TamStrcDetalleDirectorio, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Tamaño Struct Inodo:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.TamStrcInodo, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Tamaño Struct Bloques:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.TamStrcBloque, 10) + "</td></tr>\n"
	// Primer Bit Libre
	cDot += "<tr><td width=\"120\" align=\"left\">Primer Bit Libre AVD:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.PrimerBitLibreArbolDir, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Primer Bit Libre Detalle Dir:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.PrimerBitLibreDetalleDir, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Primer Bit Tabla Inodo:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.PrimerBitLibreTablaInodo, 10) + "</td></tr>\n"
	cDot += "<tr><td width=\"120\" align=\"left\">Primer Bit Bloques:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.PrimerBitLibreBloques, 10) + "</td></tr>\n"
	// Numero Mágico
	cDot += "<tr><td width=\"120\" align=\"left\">Numero Magico:</td><td width=\"130\" align=\"left\">" + strconv.FormatInt(superBoot.NumeroMagico, 10) + "</td></tr>\n"
	// Finaliza la escritura de atributos SB
	cDot += "</table>\n>];}"
	// Se escribirá el archivo dot
	state := WriteFile(p, n, "dot", cDot)
	if state {
		fmt.Println("> DOT escrito exitosamente...")
		outType := "-T" + e
		pathDot := p + n + ".dot"
		pathRep := path
		DotGenerator(outType, pathDot, pathRep)
	}
}

//SeparatePath separa el path, el nombre y la extension
func SeparatePath(path string) (string, string, string) {
	ini := 0
	eini := 0
	name := ""
	newPath := ""
	ext := ""
	pathLen := len(path)
	for i := pathLen - 1; i >= 0; i-- {
		if path[i] == '/' {
			ini = i
			break
		}
	}
	//Recolectar el nombre
	for x := ini + 1; x < pathLen; x++ {
		if path[x] == '.' {
			eini = x
			break
		}
		name += string(path[x])
	}
	//Recolectar el path
	for x := 0; x <= ini; x++ {
		newPath += string(path[x])
	}
	//Recolectar la extension
	for x := eini + 1; x < len(path); x++ {
		ext += string(path[x])
	}
	return newPath, name, ext
}

//GetString retorna una cadena
func GetString(b [16]byte) string {
	var str string
	for i := 0; i < 16; i++ {
		if unicode.IsLetter(rune(b[i])) || unicode.IsDigit(rune(b[i])) || unicode.IsSymbol(rune(b[i])) || b[i] == '/' || b[i] == '_' || b[i] == ' ' {
			str += string(b[i])
			continue
		}
		break
	}
	return str
}

//GetDateAsString : obtener la fecha como una cadena
func GetDateAsString(t Time) string {
	fyh := strconv.FormatInt(t.Day, 10) + "/" + strconv.FormatInt(t.Month, 10) + "/" + strconv.FormatInt(t.Year, 10) + " " + strconv.FormatInt(t.Hour, 10) + ":" + strconv.FormatInt(t.Minute, 10) + ":" + strconv.FormatInt(t.Seconds, 10)
	return fyh
}

//GetType retorna un string con el tipo de particion
func GetType(b byte) string {
	if b == 'l' {
		return "Logica"
	} else if b == 'e' {
		return "Extendida"
	} else {
		return "Primaria"
	}
}

//GetFit retorna un string con el fit de particion
func GetFit(b byte) string {
	if b == 'f' {
		return "Primer"
	} else if b == 'b' {
		return "Mejor"
	} else {
		return "Peor"
	}
}

//DotGenerator generara el reporte
func DotGenerator(outType string, pathDot string, pathRep string) {
	// dot -Tpng filename.dot -o outfile.png
	prc := exec.Command("dot", outType, pathDot, "-o", pathRep)
	out := bytes.NewBuffer([]byte{})
	prc.Stdout = out
	err := prc.Start()
	if err != nil {
		fmt.Println(err)
	}

	prc.Wait()

	if prc.ProcessState.Success() {
		fmt.Println("*** Reporte generado con exito. ***")
	} else {
		fmt.Println("[!] Error al compilar .dot ***")
	}
}

//--- REPORTES DE BITMAPS ---------------------------------------------------------------

// BitMapReport : genera los reportes de bitmap
func BitMapReport(path string, mp Mounted, tipo int) {
	partition := mp.Part
	contBMap := ""
	sb := ReadSuperBoot(mp.Path, partition.PartStart)
	t := getCurrentTime()
	st := GetDateAsString(t)
	position := int64(0)
	size := int64(0)
	switch tipo {
	case 1: //BitMap de Arbol de directorio
		contBMap += "Reporte BitMap Arbol de Directorios\nGenerado: " + st + "\n\n"
		position = sb.AptBmapArbolDirectorio
		size = sb.AptArbolDirectorio - position
	case 2: //BitMap de Detalle Directorio
		contBMap += "Reporte BitMap Detalle Directorio\nGenerado: " + st + "\n\n"
		position = sb.AptBmapDetalleDirectorio
		size = sb.AptDetalleDirectorio - position
	case 3: //BitMap de Tabla de Inodo
		contBMap += "Reporte BitMap Tabla de Inodos\nGenerado: " + st + "\n\n"
		position = sb.AptBmapTablaInodo
		size = sb.AptTablaInodo - position
	case 4: //BitMap de Bloque de Datos
		contBMap += "Reporte BitMap Detalle Directorios\nGenerado: " + st + "\n\n"
		position = sb.AptBmapBloques
		size = sb.AptBloques - position
	default:
		fmt.Println("[!] Codigo de reporte de BitMap incorrecto...")
		return
	}
	fmt.Println("[ I:", position, "Sz:", size, "]")
	stateBmap, bmap := GetByteArray(mp.Path, size, position)
	// Si todo sale bien, se recorre el arreglo
	if stateBmap {
		for i, b := range bmap {
			if (i % 5) == 0 {
				contBMap += " "
			}
			if (i % 50) == 0 {
				contBMap += "\n"
			}
			if b == 1 {
				contBMap += "1"
				continue
			}
			contBMap += "0"
		}
		// Escribe el archivo con la cadena de Bits
		p, n, e := SeparatePath(path)
		state := WriteFile(p, n, e, contBMap)
		if state {
			fmt.Println("*** Reporte BitMap escrito exitosamente ***")
		} else {
			fmt.Println("[!] Error al escribir Reporte BitMap...")
		}
	}
}

//WriteFile escribe un archivo
func WriteFile(path string, name string, ext string, contenido string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
		fmt.Println("Se ha creado el directorio: ", path)
	}
	f, err := os.Create(path + name + "." + ext)
	if err != nil {
		fmt.Println("[!] Error al crear archivo.")
		fmt.Println(err)
		return false
	}
	_, errw := f.WriteString(contenido)
	if err != nil {
		fmt.Println("[!] Error al escribir archivo.")
		fmt.Println(errw)
		f.Close()
		return false
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

//GetByteArray : obtiene un arreglo de Bytes
func GetByteArray(diskPath string, size int64, position int64) (bool, []byte) {
	// Abrir el Disco...
	file, err := os.Open(diskPath)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del disco
	file.Seek(position, 1)
	bmap := make([]byte, size)
	data := readBytes(file, size)
	buffer := bytes.NewBuffer(data)
	// Se decodifica y se guarda
	err = binary.Read(buffer, binary.BigEndian, &bmap)
	if err != nil {
		log.Fatal("[!] Fallo la lectura de BitMap", err)
	} else {
		fmt.Println("[ BitMap leído exitosamente ]")
		return true, bmap
	}
	return false, bmap
}

//--- REPORTE DE DIRECTORIOS ------------------------------------------------------------

// DirsReport : reporte de los direcotorios del sistema
func DirsReport(path string, pm Mounted) {
	p, n, e := SeparatePath(path)
	// Obtener Fecha y Hora actual
	ct := getCurrentTime()
	// Recuperar el SuperBoot
	sb := ReadSuperBoot(pm.Path, pm.Part.PartStart)
	// Obtener el directorio raíz
	rootDir := ReadAVD(pm.Path, sb.AptArbolDirectorio)
	// Generar contenido Dot
	codir := int64(1)
	cDot := "digraph {\nedge[arrowhead=vee]\n"
	cDot += "labelloc=\"t\";\nlabel=<REPORTE DE DIRECTORIOS<BR /><FONT POINT-SIZE=\"10\">Generado:" + GetDateAsString(ct) + "</FONT><BR /><BR />>;"
	cDot += "d1 [label=<" + GetString(rootDir.NombreDirectorio) + "<BR /><FONT POINT-SIZE=\"6\">" + GetDateAsString(rootDir.FechaCreacion) + "</FONT>> shape=folder style=filled fillcolor=darkgoldenrod1 fontsize=11];\n"
	// Creo los apuntadores a subdirectorios ...
	for i, apt := range rootDir.AptArregloSubDir {
		if apt > 0 {
			cDot += "aptr" + strconv.Itoa(int(codir)) + "" + strconv.Itoa(int(apt)) + "[label=\"[Apt" + strconv.Itoa(int(i+1)) + "]: " + strconv.Itoa(int(apt)) + "\" fillcolor=white fontcolor=black fontname=\"Helvetica\" shape=plaintext height=0.1 width=0.1 fontsize=8];\n"
		}
	}
	// Apunto la carpeta padre a los subdirectorios
	for _, apt := range rootDir.AptArregloSubDir {
		if apt > 0 {
			cDot += "d" + strconv.Itoa(int(codir)) + "->" + "aptr" + strconv.Itoa(int(codir)) + "" + strconv.Itoa(int(apt)) + " [dir=none]\n"
		}
	}
	for _, apt := range rootDir.AptArregloSubDir {
		if apt > 0 {
			position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (apt - 1))
			subdir := ReadAVD(pm.Path, position)
			cDot += CreateDirDot(subdir, codir, apt, pm.Path, sb.TamStrcArbolDirectorio, sb.AptArbolDirectorio)
		}
	}

	// Si existe creo el apuntador indirecto
	if rootDir.AptIndirecto > 0 {
		cDot += "aptr" + strconv.Itoa(int(codir)) + "" + strconv.Itoa(int(rootDir.AptIndirecto)) + "[label=\"[IND]: " + strconv.Itoa(int(rootDir.AptIndirecto)) + "\" fillcolor=white fontcolor=black fontname=\"Helvetica\" shape=plaintext height=0.1 width=0.1 fontsize=8];\n"

		cDot += "d" + strconv.Itoa(int(codir)) + "->" + "aptr" + strconv.Itoa(int(codir)) + "" + strconv.Itoa(int(rootDir.AptIndirecto)) + " [dir=none]\n"

		// Generar el codigo del directorio indirecto
		position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (rootDir.AptIndirecto - 1))
		subdir := ReadAVD(pm.Path, position)
		cDot += CreateDirDot(subdir, codir, rootDir.AptIndirecto, pm.Path, sb.TamStrcArbolDirectorio, sb.AptArbolDirectorio)
	}

	cDot += "}"
	// Se escribirá el archivo dot
	state := WriteFile(p, n, "dot", cDot)
	if state {
		fmt.Println("> DOT escrito exitosamente...")
		outType := "-T" + e
		pathDot := p + n + ".dot"
		pathRep := path
		DotGenerator(outType, pathDot, pathRep)
	}
}

// CreateDirDot : crea el codigo dot del directorio
func CreateDirDot(dir ArbolVirtualDir, padre int64, act int64, path string, sizeAVD int64, iniAVD int64) string {
	cDot := ""
	cDot += "d" + strconv.Itoa(int(act)) + "[label=<" + GetString(dir.NombreDirectorio) + "<BR /><FONT POINT-SIZE=\"6\">" + GetDateAsString(dir.FechaCreacion) + "</FONT>> shape=folder style=filled fillcolor=darkgoldenrod1 fontsize=11];\n"
	// Creacion del apuntador padre al hijo
	cDot += "aptr" + strconv.Itoa(int(padre)) + "" + strconv.Itoa(int(act)) + " -> " + "d" + strconv.Itoa(int(act)) + "\n"
	// Creo los apuntadores a subdirectorios ...
	for i, apt := range dir.AptArregloSubDir {
		if apt > 0 {
			cDot += "aptr" + strconv.Itoa(int(act)) + "" + strconv.Itoa(int(apt)) + "[label=\"[Apt" + strconv.Itoa(int(i+1)) + "]: " + strconv.Itoa(int(apt)) + "\" fillcolor=white fontcolor=black fontname=\"Helvetica\" shape=plaintext height=0.1 width=0.1 fontsize=8];\n"
		}
	}
	// Apunto la carpeta padre a los subdirectorios
	for _, apt := range dir.AptArregloSubDir {
		if apt > 0 {
			cDot += "d" + strconv.Itoa(int(act)) + "->" + "aptr" + strconv.Itoa(int(act)) + "" + strconv.Itoa(int(apt)) + " [dir=none]\n"
		}
	}
	// Si existe creo el apuntador indirecto
	if dir.AptIndirecto > 0 {
		cDot += "aptr" + strconv.Itoa(int(act)) + "" + strconv.Itoa(int(dir.AptIndirecto)) + "[label=\"[IND]: " + strconv.Itoa(int(dir.AptIndirecto)) + "\" fillcolor=white fontcolor=black fontname=\"Helvetica\" shape=plaintext height=0.1 width=0.1 fontsize=8];\n"

		cDot += "d" + strconv.Itoa(int(act)) + "->" + "aptr" + strconv.Itoa(int(act)) + "" + strconv.Itoa(int(dir.AptIndirecto)) + " [dir=none]\n"

		// Generar el codigo del directorio indirecto
		position := iniAVD + (sizeAVD * (dir.AptIndirecto - 1))
		subdir := ReadAVD(path, position)
		cDot += CreateDirDot(subdir, act, dir.AptIndirecto, path, sizeAVD, iniAVD)
	}
	for _, apt := range dir.AptArregloSubDir {
		if apt > 0 {
			position := iniAVD + (sizeAVD * (apt - 1))
			subdir := ReadAVD(path, position)
			cDot += CreateDirDot(subdir, act, apt, path, sizeAVD, iniAVD)
		}
	}
	return cDot
}
