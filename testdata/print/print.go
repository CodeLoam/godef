package print

import (
	"github.com/rogpeppe/godef/a"
	"github.com/rogpeppe/godef/b"
)

func Bar() {
	a.Stuff() //@mark(PrintStuff, "Stuff")
	//@godefPrint(PrintStuff, "json", re`{"filename":".*godef[/\\]a[/\\]a\.go","line":\d,"column":\d}\n$`)
	//@godefPrint(PrintStuff, "type", re`^.*godef[/\\]a[/\\]a\.go:\d:\d\n.*Stuff func\(\)\s*$`)
	//@godefPrint(PrintStuff, "public", re`^.*godef[/\\]a[/\\]a\.go:\d:\d\n.*Stuff func\(\)\s*$`)
	//@godefPrint(PrintStuff, "all", re`^.*godef[/\\]a[/\\]a\.go:\d:\d\n.*Stuff func\(\)\s*$`)

	var s1 b.S1 //@mark(PrintS1, "S1")
	//@godefPrint(PrintS1, "json", re`{"filename":".*godef[/\\]b[/\\]b\.go","line":\d,"column":\d}\n$`)
	//@godefPrint(PrintS1, "type", re`^.*godef[/\\]b[/\\]b\.go:\d:\d\ntype S1 struct \{\s*F1\s*int\s*f2\s*int\s*f3\s*S2\s*S2\s*\}\s*$`)
	// the succeeds, but lists no fields which seems wrong
	//@godefPrint(PrintS1, "public", re`^.*godef[/\\]b[/\\]b\.go:\d:\d\ntype S1 struct \{\s*F1\s*int\s*f2\s*int\s*f3\s*S2\s*S2\s*\}\s*$`)
	// the following fails, but it lists F1 string which seems wrong
	//@_godefPrint(PrintS1, "all", re`^.*godef[/\\]b[/\\]b\.go:\d:\d\ntype S1 struct \{\s*F1\s*int\s*f2\s*int\s*f3\s*S2\s*S2\s*S2\s*\}\s*F1\s*int\s*f2\s*int\s*f3\s*S2\s*S2\s*$`)
}
