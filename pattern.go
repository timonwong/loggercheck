package loggercheck

import (
	"bufio"
	"errors"
	"fmt"
	"go/types"
	"io"
	"sort"
	"strings"

	"github.com/timonwong/loggercheck/internal/bytebufferpool"
)

var errInvalidPattern = errors.New("invalid pattern")

type PatternGroupList []PatternGroup

func (l PatternGroupList) HasName(name string) bool {
	for _, pg := range l {
		if pg.Name == name {
			return true
		}
	}
	return false
}

func (l PatternGroupList) Names() []string {
	keys := make([]string, len(l))
	visited := make(map[string]struct{})
	for i, pg := range l {
		if _, ok := visited[pg.Name]; ok {
			continue
		}
		visited[pg.Name] = struct{}{}
		keys[i] = pg.Name
	}
	sort.Strings(keys)
	return keys
}

func parsePatternFile(r io.Reader) (result []PatternGroup, err error) {
	scanner := bufio.NewScanner(r)
	var lineCnt int

	patternsByImport := make(map[string][]Pattern)
	for scanner.Scan() {
		lineCnt++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") { // comments
			continue
		}

		pat, err := parsePattern(line)
		if err != nil {
			return nil, fmt.Errorf("error parse pattern at line %d: %w", lineCnt, err)
		}
		patternsByImport[pat.PackageImport] = append(patternsByImport[pat.PackageImport], pat)
	}

	for packageImport, patterns := range patternsByImport {
		result = append(result, PatternGroup{
			Name:          "custom", // NOTE(timonwong) Always "custom" for external patterns
			PackageImport: packageImport,
			Patterns:      patterns,
		})
	}
	return result, nil
}

type PatternGroup struct {
	Name          string
	PackageImport string
	Patterns      []Pattern
}

func (g *PatternGroup) Match(fn *types.Func, pkg *types.Package) bool {
	pkgPath := pkg.Path()
	if pkgPath != g.PackageImport && !strings.HasSuffix(pkgPath, "/vendor/"+g.PackageImport) {
		return false
	}

	sig := fn.Type().(*types.Signature) // it's safe since we already checked
	for _, pattern := range g.Patterns {
		if pattern.match(fn, sig) {
			return true
		}
	}

	return false
}

func emptyQualifier(*types.Package) string {
	return ""
}

type Pattern struct {
	PackageImport string
	ReceiverType  string
	FuncName      string
	IsReceiver    bool
}

func (p *Pattern) match(fn *types.Func, sig *types.Signature) bool {
	// we do not check package import here since it's already checked
	if fn.Name() != p.FuncName {
		return false
	}

	recv := sig.Recv()
	isReceiver := recv != nil
	if isReceiver != p.IsReceiver {
		return false
	}

	if isReceiver {
		recvType := recv.Type()
		recvTypeBuf := bytebufferpool.Get()
		defer bytebufferpool.Put(recvTypeBuf)
		types.WriteType(recvTypeBuf, recvType, emptyQualifier)
		if recvTypeBuf.String() != p.ReceiverType {
			return false
		}
	}

	return true
}

func parsePattern(s string) (pat Pattern, err error) {
	lastDot := strings.LastIndexFunc(s, func(r rune) bool {
		return r == '.' || r == '/'
	})
	if lastDot == -1 || s[lastDot] == '/' {
		return Pattern{}, errInvalidPattern
	}

	importOrReceiver := s[:lastDot]
	pat.FuncName = s[lastDot+1:]

	if strings.HasPrefix(s, "(") { // package
		if !strings.HasSuffix(importOrReceiver, ")") {
			return Pattern{}, errInvalidPattern
		}

		var isPointerReceiver bool
		pat.IsReceiver = true
		receiver := importOrReceiver[1 : len(importOrReceiver)-1]
		if strings.HasPrefix(receiver, "*") {
			isPointerReceiver = true
			receiver = receiver[1:]
		}

		typeDotIdx := strings.LastIndexFunc(receiver, func(r rune) bool {
			return r == '.' || r == '/'
		})
		if typeDotIdx == -1 || receiver[typeDotIdx] == '/' {
			return Pattern{}, errInvalidPattern
		}
		receiverType := receiver[typeDotIdx+1:]
		if isPointerReceiver {
			receiverType = "*" + receiverType
		}
		pat.ReceiverType = receiverType
		pat.PackageImport = receiver[:typeDotIdx]
	} else {
		pat.PackageImport = importOrReceiver
	}

	return pat, nil
}
