package language

import "regexp"

var packageDeclRE = regexp.MustCompile(`(?m)^\s*package\s+"[^"]+"\s*$`)

func HasPackageDeclaration(source string) bool { return packageDeclRE.FindStringIndex(source) != nil }
