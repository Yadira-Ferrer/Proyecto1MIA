package main

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	initState       = 0
	idState         = 1
	paramOrNumState = 2
	numberState     = 3
	paramState      = 4
	commentState    = 5
	stringState     = 6
	pathState       = 7
	finalState      = 8
	errorState      = -1
)

//Token para el análisis de la entrada...
type Token struct {
	name  string
	value string
}

func analizar(entrada string) []Token {
	fmt.Println("\n===== INICIA EL ANALISIS =======================================")
	state := initState
	entrada = entrada + " \n$"
	ctoken := ""
	tokenList := make([]Token, 0, 10)
	inputLen := len(entrada)

	for i := 0; i < inputLen; i++ {
		currentChar := entrada[i]
		switch state {
		case initState:
			if unicode.IsLetter(rune(currentChar)) {
				state = idState
				ctoken += string(currentChar)
			} else if currentChar == '-' {
				state = paramOrNumState
				ctoken += string(currentChar)
			} else if unicode.IsDigit(rune(currentChar)) {
				state = numberState
				ctoken += string(currentChar)
			} else if currentChar == '#' {
				state = commentState
				ctoken += string(currentChar)
			} else if currentChar == '"' {
				state = stringState
				ctoken += string(currentChar)
			} else if currentChar == '/' {
				state = pathState
				ctoken += string(currentChar)
			} else if currentChar == '$' {
				fmt.Println("[*] Se encontraron ", len(tokenList), " TOKENS.")
			} else if unicode.IsSpace(rune(currentChar)) {
				/* Se ignoran */
				state = initState
				ctoken = ""
			} else {
				fmt.Println("[!~SCANNER] Caracter que no hizo match: ", string(currentChar))
				state = errorState
			}
		case idState:
			if unicode.IsLetter(rune(currentChar)) || unicode.IsDigit(rune(currentChar)) || currentChar == '_' || currentChar == '.' {
				ctoken += string(currentChar)
			} else {
				tokentype := getIDType(ctoken)
				tokenList = append(tokenList, Token{tokentype, ctoken})
				ctoken = ""
				state = initState
				i = i - 1
			}
		case paramOrNumState:
			if unicode.IsDigit(rune(currentChar)) {
				state = numberState
				ctoken += string(currentChar)
			} else if unicode.IsLetter(rune(currentChar)) {
				ctoken = ""
				state = paramState
				ctoken += string(currentChar)
			} else if currentChar == '>' {
				/* Viene -> token asignación */
				ctoken = ""
				state = initState
			} else {
				/* Algo que no debía venir... */
				state = errorState
			}
		case numberState:
			if unicode.IsDigit(rune(currentChar)) {
				ctoken += string(currentChar)
			} else {
				tokenList = append(tokenList, Token{"numero", ctoken})
				ctoken = ""
				state = initState
				i = i - 1
			}
		case paramState:
			if unicode.IsLetter(rune(currentChar)) {
				ctoken += string(currentChar)
			} else {
				tokenList = append(tokenList, Token{"parametro", ctoken})
				ctoken = ""
				state = initState
				i = i - 1
			}
		case commentState:
			if currentChar != '\n' {
				ctoken += string(currentChar)
			} else {
				tokenList = append(tokenList, Token{"comentario", ctoken})
				ctoken = ""
				state = initState
			}
		case stringState:
			if currentChar != '"' {
				ctoken += string(currentChar)
			} else {
				ctoken += string(currentChar)
				tokenList = append(tokenList, Token{"cadena", ctoken})
				ctoken = ""
				state = initState
			}
		case pathState:
			if currentChar != ' ' && currentChar != '\n' {
				ctoken += string(currentChar)
			} else {
				tokenList = append(tokenList, Token{"path", ctoken})
				ctoken = ""
				state = initState
			}
		default:
			fmt.Println("[!~SCANNER] Caracter ", string(currentChar), " genera error.")
			fmt.Println("===== FINALIZA EL ANALISIS =====================================")
			return make([]Token, 0, 1)
		}
	}
	fmt.Println("===== FINALIZA EL ANALISIS =====================================")
	return tokenList
}

func getIDType(token string) string {
	tk := strings.ToLower(token)
	switch tk {
	case
		"exec",
		"pause",
		"mkdisk",
		"rmdisk",
		"fdisk",
		"mount",
		"unmount",
		"rep":
		return "comando"
	}
	return "id"
}
