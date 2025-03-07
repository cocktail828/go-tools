// Package morse implements morse encoding, fork from https://github.com/martinlindhe/morse
package morse

import (
	"errors"
	"strings"

	"github.com/cocktail828/go-tools/z/reflectx"
)

var (
	morseLetter = map[string]string{
		"a":  ".-",
		"b":  "-...",
		"c":  "-.-.",
		"d":  "-..",
		"e":  ".",
		"f":  "..-.",
		"g":  "--.",
		"h":  "....",
		"i":  "..",
		"j":  ".---",
		"k":  "-.-",
		"l":  ".-..",
		"m":  "--",
		"n":  "-.",
		"o":  "---",
		"p":  ".--.",
		"q":  "--.-",
		"r":  ".-.",
		"s":  "...",
		"t":  "-",
		"u":  "..-",
		"v":  "...-",
		"w":  ".--",
		"x":  "-..-",
		"y":  "-.--",
		"z":  "--..",
		"ä":  ".-.-",
		"å":  ".-.-",
		"ç":  "-.-..",
		"ĉ":  "-.-..",
		"ö":  "-.-..",
		"ø":  "---.",
		"ð":  "..--.",
		"ü":  "..--",
		"ŭ":  "..--",
		"ch": "----",
		"0":  "-----",
		"1":  ".----",
		"2":  "..---",
		"3":  "...--",
		"4":  "....-",
		"5":  ".....",
		"6":  "-....",
		"7":  "--...",
		"8":  "---..",
		"9":  "----.",
		".":  ".-.-.-",
		",":  "--..--",
		"`":  ".----.",
		"?":  "..--..",
		"!":  "..--.",
		":":  "---...",
		";":  "-.-.-",
		"\"": ".-..-.",
		"'":  ".----.",
		"=":  "-...-",
		"(":  "-.--.",
		")":  "-.--.-",
		"$":  "...-..-",
		"&":  ".-...",
		"@":  ".--.-.",
		"+":  ".-.-.",
		"-":  "-....-",
		"/":  "-..-.",
	}
)

// Encode encodes clear text using `alphabet` mapping
func Encode(b []byte, separator string) (dst string, err error) {
	s := strings.ToLower(reflectx.BytesToString(b))
	if strings.Contains(s, " ") {
		return dst, errors.New("can't contain spaces")
	}
	for _, letter := range s {
		if let := string(letter); morseLetter[let] != "" {
			dst += morseLetter[let] + separator
		} else {
			return dst, errors.New("unexpect char: " + let)
		}
	}
	dst = strings.Trim(dst, separator)
	return
}

// Decode decodes morse code using `alphabet` mapping
func Decode(b []byte, separator string) (dst string, err error) {
	for _, part := range strings.Split(reflectx.BytesToString(b), separator) {
		found := false
		for key, letter := range morseLetter {
			if letter == part {
				dst += key
				found = true
				break
			}
		}
		if !found {
			return dst, errors.New("unknown character: " + part)
		}
	}
	return
}
