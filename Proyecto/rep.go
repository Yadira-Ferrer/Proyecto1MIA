package main

import (
	"bytes"
	"fmt"
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
	state := WriteDotFile(p, n, cdot)
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
	state := WriteDotFile(p, n, cdot)
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
	state := WriteDotFile(p, n, cDot)
	if state {
		fmt.Println("> DOT escrito exitosamente...")
		outType := "-T" + e
		pathDot := p + n + ".dot"
		pathRep := path
		DotGenerator(outType, pathDot, pathRep)
	}
}

//WriteDotFile escribe el archivo .dot
func WriteDotFile(path string, name string, contenido string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
		fmt.Println("Se ha creado el directorio: ", path)
	}
	f, err := os.Create(path + name + ".dot")
	if err != nil {
		fmt.Println("[!] Error al crear archivo Dot.")
		fmt.Println(err)
		return false
	}
	_, errw := f.WriteString(contenido)
	if err != nil {
		fmt.Println("[!] Error al escribir archivo Dot.")
		fmt.Println(errw)
		f.Close()
		return false
	}
	//fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
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
		if unicode.IsLetter(rune(b[i])) || unicode.IsDigit(rune(b[i])) || unicode.IsSymbol(rune(b[i])) {
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
		fmt.Println("*** Reporte generado con exito ***")
		fmt.Println(out.String())
	}
}
