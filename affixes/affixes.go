//	This file is part of Fwew.
//	Fwew is free software: you can redistribute it and/or modify
// 	it under the terms of the GNU General Public License as published by
// 	the Free Software Foundation, either version 3 of the License, or
// 	(at your option) any later version.
//
//	Fwew is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//
//	You should have received a copy of the GNU General Public License
//	along with Fwew.  If not, see http://gnu.org/licenses/

// Package affixes handles affix parsing of input
package affixes

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tirea/fwew/util"
)

var debug = false

// Word is a struct that contains all the data about a given word
type Word struct {
	ID             string
	LangCode       string
	Navi           string
	Target         string
	Attempt        string
	IPA            string
	InfixLocations string
	PartOfSpeech   string
	Definition     string
	Affixes        map[string][]string
}

func (w Word) String() string {
	// this string only doesn't get translated or called from util.Text() because they're var names
	return fmt.Sprintf("Id: %s\nLangCode: %s\nNavi: %s\nTarget: %s\nAttempt: %s\nIPA: %s\nInfixLocations: %s\nPartOfSpeech: %s\nDefinition: %s\nAffixes: %v\n",
		w.ID, w.LangCode, w.Navi, w.Target, w.Attempt, w.IPA, w.InfixLocations, w.PartOfSpeech, w.Definition, w.Affixes)
}

// InitWordStruct is basically a constructer for Word struct
func InitWordStruct(w Word, dataFields []string) Word {
	const (
		idField  int = 0 // dictionary.tsv line Field 0 is Database ID
		lcField  int = 1 // dictionary.tsv line field 1 is Language Code
		navField int = 2 // dictionary.tsv line field 2 is Na'vi word
		ipaField int = 3 // dictionary.tsv line field 3 is IPA data
		infField int = 4 // dictionary.tsv line field 4 is Infix location data
		posField int = 5 // dictionary.tsv line field 5 is Part of Speech data
		defField int = 6 // dictionary.tsv line field 6 is Local definition
	)
	w.ID = dataFields[idField]
	w.LangCode = dataFields[lcField]
	w.Navi = dataFields[navField]
	w.IPA = dataFields[ipaField]
	w.InfixLocations = dataFields[infField]
	w.PartOfSpeech = dataFields[posField]
	w.Definition = dataFields[defField]
	w.Affixes = map[string][]string{}

	return w
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func contains(s []string, q []string) bool {
	if len(q) == 0 || len(s) == 0 {
		return false
	}
	// search for any instance of a thing in q...
	for _, x := range q {
		// ... that exists in s.
		for _, y := range s {
			if y == x {
				return true
			}
		}
	}

	return false
}

func containsStr(s []string, q string) bool {
	if len(q) == 0 || len(s) == 0 {
		return false
	}
	for _, x := range s {
		if q == x {
			return true
		}
	}
	return false
}

func index(s []string, q string) (int, error) {
	for i, v := range s {
		if v == q {
			return i, nil
		}
	}
	return -1, errors.New(util.Text("strNotInSliceError"))
}

func prefix(w Word) Word {
	var re *regexp.Regexp
	var reString string
	var attempt string
	var lenPre = []string{"pe", "fray", "tsay", "fay", "pay", "ay", "me", "pxe"}
	var matchPrefixes = []string{}

	switch w.PartOfSpeech {
	case "n.", "pn.":
		reString = "(pe)?(fray)?(tsay)?(fay)?(pay)?(ay)?(fra)?(fì)?(tsa)?(me)?(pxe)?(fne)?(munsna)?"
	case "dem.", "dem., pn.":
		reString = "(pe)?(fray)?(tsay)?(fay)?(pay)?(ay)?(fra)?(me)?(pxe)?(fne)?(munsna)?"
	case "adj.":
		reString = "(nì|a)?"
	case "vin.", "vtr.", "vim.", "vtrm.", "v.":
		reString = "(ketsuk|tsuk)?"
	default:
		return w // Not a type that has a prefix, return word without attempting.
	}

	reString = reString + ".+"
	if debug {
		fmt.Printf("Prefix reString: %s\n", reString)
	}
	re = regexp.MustCompile(reString)
	tmp := re.FindAllStringSubmatch(w.Target, -1)
	if len(tmp) > 0 && len(tmp[0]) >= 1 {
		matchPrefixes = tmp[0][1:]
	}
	matchPrefixes = deleteEmpty(matchPrefixes)
	if debug {
		fmt.Printf("matchPrefixes: %s\n", matchPrefixes)
	}

	// no productive prefixes found; why bother to continue?
	if len(matchPrefixes) == 0 {
		return w
	}

	// build what prefixes to put on
	for _, p := range matchPrefixes {
		attempt = attempt + p
	}

	// check for leniting prefix
	if contains(matchPrefixes, lenPre) {
		// lenite first
		w = lenite(w)
		// then add prefixes
		if w.Attempt != w.Navi {
			// leniting prefix, lenition occured
			w.Attempt = attempt + w.Attempt
		} else {
			// leniting prefix, lenition did not occur
			w.Attempt = attempt + w.Navi
		}
	} else {
		// otherwise just add the prefixes to create the attempt
		w.Attempt = attempt + w.Attempt
	}

	w.Affixes["prefixes"] = matchPrefixes

	//prodGerundAffix := []string{"tì", "us"}
	//prodActPartAffixPre := []string{"a", "us"}
	//prodPassPartAffixPre := []string{"a", "awn"}
	return w
}

func suffix(w Word) Word {
	var re *regexp.Regexp
	var reString string
	var attempt string
	var matchSuffixes = []string{}

	switch w.PartOfSpeech {
	case "n.", "pn.", "dem.", "dem., pn.":
		reString = "(nga')?(tsyìp)?(o)?(pe)?(ìri)?(ìlä)?(ìl)?(eyä)?(yä)?(ä)?(it)?(ri)?(ru)?(ti)?(tu)?(ur)?(l)?(r)?(t)?(y)?"
		reString += "(mungwrr)?(kxamlä)?(tafkip)?(pxisre)?(pximaw)?(ftumfa)?(mìkam)?(nemfa)?(takip)?(lisre)?(talun)?"
		reString += "(krrka)?(teri)?(fkip)?(pxaw)?(pxel)?(luke)?(rofa)?(fpi)?(ftu)?(kip)?(vay)?(lok)?(maw)?"
		reString += "(sìn)?(sre)?(few)?(kam)?(kay)?(nuä)?(sko)?(yoa)?(äo)?(eo)?(fa)?(hu)?(ka)?(mì)?(na)?(ne)?(ta)?(io)?(uo)?(ro)?(wä)?"
	case "adj.":
		reString = "(a)?"
	case "num.":
		reString = "(ve)?(a)?"
	case "v.", "vin.", "vtr.", "vim.", "vtrm.", "svin.":
		inf := w.Affixes["infixes"]
		if len(inf) > 0 && (inf[0] == "us" || inf[0] == "awn") {
			reString = "(a)?"
		} else {
			reString = "(tswo|yu)?"
		}
	default:
		return w // Not a type that has a suffix, return word without attempting.
	}

	reString = w.Attempt + reString
	if debug {
		fmt.Printf("Suffix reString: %s\n", reString)
	}
	re = regexp.MustCompile(reString)
	tmp := re.FindAllStringSubmatch(w.Target, -1)
	if len(tmp) > 0 && len(tmp[0]) >= 1 {
		matchSuffixes = tmp[0][1:]
	}
	matchSuffixes = deleteEmpty(matchSuffixes)
	if debug {
		fmt.Printf("matchSuffixes: %s\n", matchSuffixes)
	}

	// no productive prefixes found; why bother to continue?
	if len(matchSuffixes) == 0 {
		return w
	}

	// build what prefixes to put on
	for _, p := range matchSuffixes {
		attempt = attempt + p
	}

	w.Attempt = w.Attempt + attempt
	w.Affixes["suffixes"] = matchSuffixes
	return w
}

func infix(w Word) Word {
	// Have we already attempted infixes?
	if _, ok := w.Affixes["infixes"]; ok {
		return w
	}

	var re *regexp.Regexp
	var reString string
	var attempt string
	var pos0InfixRe = "(äp)?(eyk)?"
	var pos1InfixRe = "(ìyev|iyev|ìlm|ìly|ìrm|ìry|ìsy|alm|aly|arm|ary|asy|ìm|imv|irv|ìy|am|ay|er|iv|ol)?"
	var pos2InfixRe = "(eiy|ei|äng|eng|ats|uy)?"
	var pos0InfixString string
	var pos1InfixString string
	var pos2InfixString string
	var matchInfixes = []string{}

	reString = strings.Replace(w.InfixLocations, "<1>", pos0InfixRe, 1)
	reString = strings.Replace(reString, "<2>", pos1InfixRe, 1)
	reString = strings.Replace(reString, "<3>", pos2InfixRe, 1)
	if debug {
		fmt.Printf("Infix reString: %s\n", reString)
	}

	re = regexp.MustCompile(reString)
	tmp := re.FindAllStringSubmatch(w.Target, -1)
	if len(tmp) > 0 && len(tmp[0]) >= 1 {
		matchInfixes = tmp[0][1:]
	}
	matchInfixes = deleteEmpty(matchInfixes)

	for _, i := range matchInfixes {
		if i == "äp" || i == "eyk" {
			pos0InfixString = pos0InfixString + i
		} else if i == "eiy" || i == "ei" || i == "äng" || i == "eng" || i == "ats" || i == "uy" {
			pos2InfixString = i
		} else {
			pos1InfixString = i
		}
	}

	attempt = strings.Replace(w.InfixLocations, "<1>", pos0InfixString, 1)
	attempt = strings.Replace(attempt, "<2>", pos1InfixString, 1)
	attempt = strings.Replace(attempt, "<3>", pos2InfixString, 1)

	/*
		hardCodeHax := map[string][]string{}
		hardCodeHax["poltxe"] = []string{"plltxe", "ol"}
		hardCodeHax["molte"] = []string{"mllte", "ol"}
	*/

	if containsStr(matchInfixes, "eiy") {
		eiy, _ := index(matchInfixes, "eiy")
		matchInfixes[eiy] = "ei"
	}
	if debug {
		fmt.Printf("matchInfixes: %s\n", matchInfixes)
	}

	w.Attempt = attempt
	if len(matchInfixes) != 0 {
		w.Affixes["infixes"] = matchInfixes
	}

	return w
}

func lenite(w Word) Word {
	// Have we already attempted lenition?
	if _, ok := w.Affixes["lenition"]; ok {
		return w
	}
	switch {
	case strings.HasPrefix(w.Navi, "kx"):
		w.Attempt = strings.Replace(w.Attempt, "kx", "k", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "kx->k")
		return w
	case strings.HasPrefix(w.Navi, "px"):
		w.Attempt = strings.Replace(w.Attempt, "px", "p", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "px->p")
		return w
	case strings.HasPrefix(w.Navi, "tx"):
		w.Attempt = strings.Replace(w.Attempt, "tx", "t", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "tx->t")
		return w
	case strings.HasPrefix(w.Navi, "k"):
		w.Attempt = strings.Replace(w.Attempt, "k", "h", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "k->h")
		return w
	case strings.HasPrefix(w.Navi, "p"):
		w.Attempt = strings.Replace(w.Attempt, "p", "f", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "p->f")
		return w
	case strings.HasPrefix(w.Navi, "ts"):
		w.Attempt = strings.Replace(w.Attempt, "ts", "s", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "ts->s")
		return w
	case strings.HasPrefix(w.Navi, "t"):
		w.Attempt = strings.Replace(w.Attempt, "t", "s", 1)
		w.Affixes["lenition"] = append(w.Affixes["lenition"], "t->s")
		return w
	case strings.HasPrefix(w.Navi, "'"):
		if !strings.HasPrefix(w.Target, "'") {
			w.Attempt = strings.Replace(w.Attempt, "'", "", 1)
			w.Affixes["lenition"] = append(w.Affixes["lenition"], "'->")
			return w
		}
		return w
	default:
		return w
	}
}

// Reconstruct is the main function of affixes.go, responsible for the affixing algorithm
func Reconstruct(w Word) Word {

	w.Attempt = w.Navi

	// only try to infix verbs, si has empty string as part of speech
	if containsStr([]string{"vin.", "vtr.", "vim.", "vtrm.", "v.", "svin.", "vin., svin."}, w.PartOfSpeech) ||
		w.PartOfSpeech == "" {
		w = infix(w)
		if debug {
			fmt.Println("INFIX")
			fmt.Printf("Navi: %s | Attempt: %s | Target: %s\n", w.Navi, w.Attempt, w.Target)
		}
		if w.Attempt == w.Target {
			return w
		}
	}

	w = prefix(w)
	if debug {
		fmt.Println("PREFIX")
		fmt.Printf("Navi: %s | Attempt: %s | Target: %s\n", w.Navi, w.Attempt, w.Target)
	}
	if w.Attempt == w.Target {
		return w
	}

	if len(w.Target) > 1 && !containsStr([]string{"px", "tx", "kx", "ts"}, w.Target[:2]) && !strings.HasPrefix(w.Target, "'") {
		w = lenite(w)
		if debug {
			fmt.Println("LENITE")
			fmt.Printf("Navi: %s | Attempt: %s | Target: %s\n", w.Navi, w.Attempt, w.Target)
		}
		if w.Attempt == w.Target {
			return w
		}
	}

	w = suffix(w)
	if debug {
		fmt.Println("SUFFIX")
		fmt.Printf("Navi: %s | Attempt: %s | Target: %s\n", w.Navi, w.Attempt, w.Target)
	}
	if w.Attempt == w.Target {
		return w
	}

	w = lenite(w)
	if debug {
		fmt.Println("LENITE")
		fmt.Printf("Navi: %s | Attempt: %s | Target: %s\n", w.Navi, w.Attempt, w.Target)
	}
	if w.Attempt == w.Target {
		return w
	}

	if debug {
		fmt.Println("GIVING UP")
	}
	return Word{ID: "-1"}
}
