package main

import "fmt"

// dumpTypes outputs the type information recorded by the parser as a C header.
func dumpTypes(p *parser) {
	// Print predeclared identifiers.
	def := p.types["bool"]
	fmt.Println(def.Def())
	// Print enums.
	for _, tag := range p.enumTags {
		t := p.enums[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print structs.
	for _, tag := range p.structTags {
		t := p.structs[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print unions.
	for _, tag := range p.unionTags {
		t := p.unions[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print typedefs.
	for _, def := range p.typedefs {
		fmt.Println(def.Def())
	}
}
