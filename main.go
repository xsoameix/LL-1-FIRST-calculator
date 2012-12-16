package main

import (
	"fmt"
)

var (
	productions map[string][][]string = map[string][][]string{
		"expr": {
			{"term", "expr'"}},
		"expr'": {
			{"add", "term", "expr'"},
			{"epsilon"}},
		"term": {
			{"factor", "term'"}},
		"term'": {
			{"multi", "factor", "term'"},
			{"epsilon"}},
		"factor": {
			{"(", "expr", ")"},
			{"digit"}},
		"add": {
			{"+"},
			{"-"}},
		"multi": {
			{"*"},
			{"/"}}}
	/*productions map[string][][]string = map[string][][]string{
	"stmt": {
		{"if-stmt", "$"},
		{"other"}},
	"if-stmt": {
		{"if", "(", "exp", ")", "stmt", "else-part"}},
	"else-part": {
		{"else", "stmt"},
		{"epsilon"}},
	"exp": {
		{"0"},
		{"1"}}}*/
	nonTerminals, terminals []string
	firsts                  map[string][][]string = make(map[string][][]string)
	follows                 map[string][]string   = make(map[string][]string)
	productionsStack        map[string][][]string = productions
	null                    map[string]bool       = make(map[string]bool)
)

func main() {
	makeNonTerminals()
	makeTerminals()
	makeNulls()
	fmt.Println(nonTerminals)
	fmt.Println(terminals)
	fmt.Println(null)

	makeFirsts()
	fmt.Println(firsts)

	makeFollows()
	fmt.Println(follows)
}

func makeNonTerminals() {
	for nonTerminal, _ := range productions {
		nonTerminals = append(nonTerminals, nonTerminal)
	}
}

func makeTerminals() {
	for _, or := range productions {
		for _, production := range or {
			for _, token := range production {
				if _, k := productions[token]; !k && token != "epsilon" {
					terminals = append(terminals, token)
				}
			}
		}
	}
}

func makeNulls() {
	for _, nonTerminal := range nonTerminals {
		if _, exist := null[nonTerminal]; exist {
			continue
		}
		null[nonTerminal] = makeNull(nonTerminal)
	}
}

func makeNull(nonTerminal string) bool {
	hasTerminal := false
	hasEpsilon := false
	productionNT := [][]string{}
	for I, production := range productions[nonTerminal] {
		productionNT = append(productionNT, []string{})
		if len(production) == 1 && production[0] == "epsilon" {
			hasEpsilon = true
		}
		for _, token := range production {
			switch {
			case contains(token, terminals):
				hasTerminal = true
			case contains(token, nonTerminals):
				productionNT[I] = append(productionNT[I], token)
			}
		}
	}
	if !hasTerminal && !hasEpsilon {
		isNull := []bool{}
		for I, production := range productionNT {
			for i, token := range production {
				if _, exist := null[token]; exist {
					if i == 0 {
						isNull = append(isNull, null[token])
					} else {
						isNull[I] = isNull[I] && null[token]
					}
					continue
				}
				if i == 0 {
					isNull = append(isNull, makeNull(token))
					null[token] = isNull[I]
				} else {
					null[token] = makeNull(token)
					isNull[I] = isNull[I] && null[token]
				}
			}
			if I >= 1 {
				isNull[I] = isNull[I] || isNull[I-1]
			}
		}
		return isNull[len(isNull)-1]
	}
	return hasEpsilon || !hasTerminal
}

func makeFirsts() {
	iOfOr := -2
	for nonTerminal, or := range productions {
		first(nonTerminal, or, iOfOr)
	}
}

/*iOfOr, len(firsts[nonT]), -1
-2 != 0 -1 add in new first
-2 != 1 -1 add in new first
-2 != 1 -1 add in new first

0 != 0 -1 add in new first
0 == 1 -1 add in old first
0 == 1 -1 add in old first

1 != 1 -1
1 == 2 -1
1 == 2 -1

2 != 2 -1
2 == 3 -1
2 == 3 -1*/

func first(nonTerminal string, or [][]string, iOfOr int) {
	for i, production := range or {
		firstToken := production[0]
		switch {
		case contains(firstToken, nonTerminals):
			first(nonTerminal, productions[firstToken], i)
		case firstToken == "epsilon": //epsilon isn't terminal, but in first
			fallthrough
		case contains(firstToken, terminals):
			if iOfOr == len(firsts[nonTerminal])-1 {
				firsts[nonTerminal][iOfOr] = append(
					firsts[nonTerminal][iOfOr], firstToken)
			} else {
				firsts[nonTerminal] = append(firsts[nonTerminal],
					[]string{firstToken})
			}
		}
	}
}

func makeFollows() {
	for nonTerminal, or := range productionsStack {
		for _, production := range or {
			for i, token := range production {
				if _, finded := follows[token]; finded || token == nonTerminal { //A != c
					continue
				}
				if contains(token, nonTerminals) {
					follows[token] = append(follows[token],
						follow(nonTerminal, token, production[i:])...)
				}
			}
		}
	}
}

func follow(nonTerminal, token string, production []string) []string {
	if contains(token, terminals) {
		return []string{token}
	}
	var follow_ []string
	if token == "expr" {
		follow_ = append(follow_, "$")
	}
	//A -> Bc, follow(B) U= first(c)
	hasTokenBehind := false
	thatIsEpsilon := false
	if len(production) >= 2 {
		hasTokenBehind = true
		rest := production[1]
		if contains(rest, terminals) {
			follow_ = append(follow_, rest)
			return follow_
		}
		for i, production := range productions[rest] {
			if Null(production) {
				thatIsEpsilon = true
			} else {
				follow_ = append(follow_, firsts[rest][i]...)
			}
		}
	}

	if hasTokenBehind && !thatIsEpsilon {
		return follow_
	}
	//there is token behind and it is epsilon
	//or there has no token behind

	//A -> B || ( A -> Bc && c -> epsilon ), A != c, follow(B) U= follow(A)
	if _, finded := follows[nonTerminal]; finded {
		follow_ = append(follow_, follows[nonTerminal]...)
		return follow_
	}
Outer:
	for nonTerminal1, or := range productionsStack {
		for _, production := range or {
			for i, token2 := range production {
				if token2 == nonTerminal && nonTerminal1 != nonTerminal { //A != c
					follows[nonTerminal] = append(follows[nonTerminal],
						follow(nonTerminal1, nonTerminal, production[i:])...)
					follow_ = append(follow_, follows[nonTerminal]...)
					break Outer
				}
			}
		}
	}
	return follow_
}

func Null(production []string) bool {
	if len(production) == 1 && production[0] == "epsilon" {
		return true
	}
	return false
}

func contains(token string, lst []string) bool {
	for _, value := range lst {
		if token == value {
			return true
		}
	}
	return false
}
