package main

import (
	"fmt"
	"stack"
	"time"
)

var (
	startSymbol = "expr"
	EOF = "$"
	productions = map[string][][]string{
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
	firstsLog               = make(map[string][]bool)
	followsLog              = make(map[string][]bool)
	followsPos              = make(map[string]map[string][]int)
	nonTerminals, terminals = Set{}, Set{}
	firsts                  = make(map[string][]Set)
	follows                 = make(map[string]Set)
	table                   = make(map[string]map[string][]string)
	null                    = make(map[string]bool)
)

type Set map[string]struct{}

func NewSet(i ...string) Set {
	s := Set{}
	for _, x := range i {
		s[x] = struct{}{}
	}
	return s
}

func (s Set) String() string {
	str := "[ "
	for i := range s {
		str += i + " "
	}
	return str + "]"
}

func (a Set) has(s string) bool {
	_, ok := a[s]
	return ok
}

func (a Set) insert(s string) {
	a[s] = struct{}{}
}

func (a Set) union(b Set) Set {
	r := Set{}
	for i := range a {
		r[i] = struct{}{}
	}
	for i := range b {
		r[i] = struct{}{}
	}
	return r
}

func main() {
	makeNonTerminals()
	makeTerminals()
	makeNulls()
	fmt.Println("nonTerminal\n", nonTerminals)
	fmt.Println("terminal\n", terminals)
	fmt.Println("Null\n", null)

	makeFirsts()
	fmt.Println("First\n", firsts)

	makeFollows()
	fmt.Println("Follows\n", follows)

	makeTable()
	fmt.Println("Table\n", table)

	show()

	parse([]string{"digit", "+", "digit", "*", "digit"})
	fmt.Println("parsing done")
}

func makeNonTerminals() {
	for nonTerminal := range productions {
		nonTerminals.insert(nonTerminal)
	}
}

func makeTerminals() {
	for nonTerminal := range nonTerminals {
		for _, production := range productions[nonTerminal] {
			for _, token := range production {
				if !nonTerminals.has(token) && token != "epsilon" {
					terminals.insert(token)
				}
			}
		}
	}
	terminals.insert(EOF)
}

func makeNulls() {
	for nonTerminal := range nonTerminals {
		if _, exist := null[nonTerminal]; exist {
			continue
		}
		null[nonTerminal] = makeNull(nonTerminal)
	}
}

func makeNull(nonTerminal string) bool {
	//A -> B|C|epsilon, Null(A) = T
	for _, production := range productions[nonTerminal] {
		if len(production) == 1 && production[0] == "epsilon" {
			return true
		}
	}
	var (
		isBorC_Null  = false
		isBandC_Null = true
	)
	//A -> B|C, Null(A) = Null(B) || Null(C)
	for _, production := range productions[nonTerminal] {
		isBandC_Null = true
		//A -> BC, Null(A) = Null(B) && Null(C)
		for _, token := range production {
			//A -> Ba|xC, since a and x, Null(A) = F
			if terminals.has(token) {
				isBandC_Null = false
				break
			}
			if _, exist := null[token]; !exist {
				null[token] = makeNull(token)
			}
			isBandC_Null = isBandC_Null && null[token]
		}
		isBorC_Null = isBorC_Null || isBandC_Null
	}
	return isBorC_Null
}

func isNull(production []string) bool {
	//A -> B|C|epsilon, Null(A) = T
	if len(production) == 1 && production[0] == "epsilon" {
		return true
	}
	isBandC_Null := true
	//A -> BC, Null(A) = Null(B) && Null(C)
	for _, token := range production {
		//A -> Ba|xC, since a and x, Null(A) = F
		if terminals.has(token) {
			isBandC_Null = false
			break
		}
		isBandC_Null = isBandC_Null && null[token]
	}
	return isBandC_Null
}

func makeFirsts() {
	//init firstsLog = {expr:{false}, expr':{false, false}...}
	for nonTerminal := range nonTerminals {
		for i := 0; i < len(productions[nonTerminal]); i++ {
			firstsLog[nonTerminal] = append(
				firstsLog[nonTerminal], false)
		}
	}
	for nonTerminal := range nonTerminals {
		for i, production := range productions[nonTerminal] {
			if !firstsLog[nonTerminal][i] {
				firsts[nonTerminal] = append(
					firsts[nonTerminal],
					first(nonTerminal, production, i))
			}
		}
	}
}

func first(nonTerminal string, production []string, I int) Set {
	if nonTerminals.has(nonTerminal) {
		firstsLog[nonTerminal][I] = true
	}
	s := NewSet()
	//first() = {}
	if len(production) == 0 {
		return s
	}
	//first(epsilon) = {}
	if len(production) == 1 && production[0] == "epsilon" {
		return s
	}
	//first(a beta) = {a}
	if terminals.has(production[0]) {
		s.insert(production[0])
		return s
	}
	//first(alpha beta) = ?
	for _, token := range production {
		//first(alpha) = first(alpha1) U ... U first(alpha n)
		for i, production := range productions[token] {
			if !firstsLog[token][i] {
				firsts[token] = append(
					firsts[token], first(token, production, i))
			}
			s = s.union(firsts[token][i])
		}
		//if Null(alpha) = F, first(alpha beta) = first(alpha)
		if !null[token] {
			return s
		}
		//if Null(alpha) = T,
		// first(alpha beta) = first(alpha) U first(beta)
	}
	return s
}

func makeFollows() {
	//make followsPos
	for nonTerminal := range nonTerminals {
		for I, production := range productions[nonTerminal] {
			for i, token := range production {
				//N -> alpha Y beta, N != Y
				if token == nonTerminal ||
					!nonTerminals.has(token) {
					continue
				}
				if _, ok := followsPos[token]; ok {
					followsPos[token][nonTerminal] = []int{I, i}
				} else {
					followsPos[token] = map[string][]int{
						nonTerminal: {I, i}}
				}
			}
		}
	}
	for token, inWhichNonTerminal := range followsPos {
		if _, exist := follows[token]; !exist {
			s := NewSet()
			for nonTerminal, pos := range inWhichNonTerminal {
				s = s.union(follow(nonTerminal, token, pos))
			}
			follows[token] = s
		}
	}
}

func follow(nonTerminal, token string, pos []int) Set {
	production := productions[nonTerminal][pos[0]]
	s := NewSet()
	if token == startSymbol {
		s.insert(EOF)
	}
	if pos[1] < len(production)-1 {
		s = s.union(first("", production[pos[1]+1:], pos[0]))
		//N -> alpha Y beta, Null(beta) = F,
		//follow(Y) U= first(beta)
		isBandC_Null := true
		//A -> BC, Null(A) = Null(B) && Null(C)
		for _, token := range production[pos[1]+1:] {
			//A -> Ba|xC, since a and x, Null(A) = F
			if terminals.has(token) {
				isBandC_Null = false
				break
			}
			isBandC_Null = isBandC_Null && null[token]
		}
		if !isBandC_Null {
			return s
		}
	}
	//N -> alpha Y beta, Null(beta) = T,
	//follow(Y) U= first(beta) U follow(N)
	if _, exist := follows[nonTerminal]; !exist {
		s2 := NewSet()
		for nonTerminal_, pos := range followsPos[nonTerminal] {
			s2 = s2.union(follow(nonTerminal_, nonTerminal, pos))
		}
		follows[nonTerminal] = s2
		return s.union(s2)
	}
	return s.union(follows[nonTerminal])
}

func makeTable() {
	for nonTerminal := range nonTerminals {
		for terminal := range terminals {
			if _, exist := table[terminal]; !exist {
				table[terminal] = map[string][]string{
					nonTerminal: []string{}}
			} else {
				table[terminal][nonTerminal] = []string{}
			}
		}
	}
	for nonTerminal, NTfirst := range firsts {
		for i, first := range NTfirst {
			s := NewSet()
			s = s.union(first)
			if isNull(productions[nonTerminal][i]) {
				s = s.union(follows[nonTerminal])
			}
			for terminal := range s {
				table[terminal][nonTerminal] = append(
					table[terminal][nonTerminal],
					productions[nonTerminal][i]...)
			}
		}
	}
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

func show() {
	for terminal := range table {
		fmt.Println("===============")
		fmt.Println(terminal)
		for nonTerminal, val := range table[terminal] {
			fmt.Printf("%-10s%-5s\n", nonTerminal, val)
		}
	}
}

func parse(s []string) {
	s = append(s, EOF)
	stk := new(stack.Stack)
	stk.Push(EOF)
	stk.Push(startSymbol)
	for {
		time.Sleep(time.Second)
		top := stk.Top()
		fmt.Println(stk)
		if top == "epsilon" {
			stk.Pop()
		}
		if terminals.has(top) {
			if top == EOF {
				break
			}
			if top == s[0] {
				stk.Pop()
				s = s[1:]
			} else {
				fmt.Println("invalid string")
				break
			}
		}
		if nonTerminals.has(top) {
			stk.Pop()
			for i := len(table[s[0]][top]) - 1; i >= 0; i-- {
				stk.Push(table[s[0]][top][i])
			}
		}
	}
}
