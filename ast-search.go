package main

/*
 Note: Some of this code is based on example code from:
 https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
*/

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DeclType int64

const (
	WaitGroup DeclType = iota
	Cond
	Once
	Mutex
	RWMutex
	Locker
	Unknown
)

func (d DeclType) String() string {
	switch d {
	case WaitGroup:
		return "WaitGroup"
	case Cond:
		return "Cond"
	case Once:
		return "Once"
	case Mutex:
		return "Mutex"
	case RWMutex:
		return "RWMutex"
	case Locker:
		return "Locker"
	case Unknown:
		return "Unknown"
	default:
		panic("Bad declaration type!")
	}
}

type Declaration struct {
	name   string
	typeof DeclType
}

func createDecl(target string, typeof DeclType) Declaration {
	return Declaration{name: target, typeof: typeof}
}

type AnalysisState struct {
	decls  map[string][]Declaration
	counts Counts
}

type Counts struct {
	waitGroupDecls   int
	condDecls        int
	onceDecls        int
	mutexDecls       int
	rwMutexDecls     int
	lockerDecls      int
	waitGroupDone    int
	waitGroupAdd     int
	waitGroupWait    int
	mutexLock        int
	mutexUnlock      int
	rwMutexLock      int
	rwMutexUnlock    int
	lockerLock       int
	lockerUnlock     int
	condLock         int
	condUnlock       int
	condWait         int
	condSignal       int
	condBroadcast    int
	condNew          int
	onceDo           int
	unknownDone      int
	unknownAdd       int
	unknownWait      int
	unknownLock      int
	unknownUnlock    int
	unknownSignal    int
	unknownBroadcast int
	unknownDo        int
}

func (s *AnalysisState) addDecl(declaration Declaration) {
	_, ok := s.decls[declaration.name]
	if ok {
		foundDecl := false
		for _, d := range s.decls[declaration.name] {
			if d.typeof == declaration.typeof {
				fmt.Printf("Found declaration for name %s and type %s\n", declaration.name, declaration.typeof.String())
				foundDecl = true
			}
		}
		if !foundDecl {
			fmt.Printf("Adding declaration for name %s and type %s\n", declaration.name, declaration.typeof.String())
			s.decls[declaration.name] = append(s.decls[declaration.name], declaration)
		}
	} else {
		fmt.Printf("Adding declaration for name %s and type %s\n", declaration.name, declaration.typeof.String())
		s.decls[declaration.name] = append(s.decls[declaration.name], declaration)
	}
}

func (s *AnalysisState) addWaitGroupDecl() {
	s.counts.waitGroupDecls++
}

func (s *AnalysisState) addCondDecl() {
	s.counts.condDecls++
}

func (s *AnalysisState) addOnceDecl() {
	s.counts.onceDecls++
}

func (s *AnalysisState) addMutexDecl() {
	s.counts.mutexDecls++
}

func (s *AnalysisState) addRWMutexDecl() {
	s.counts.rwMutexDecls++
}

func (s *AnalysisState) addLockerDecl() {
	s.counts.lockerDecls++
}

func (s *AnalysisState) addWaitGroupDone() {
	s.counts.waitGroupDone++
}

func (s *AnalysisState) addWaitGroupAdd() {
	s.counts.waitGroupAdd++
}

func (s *AnalysisState) addWaitGroupWait() {
	s.counts.waitGroupWait++
}

func (s *AnalysisState) addMutexLock() {
	s.counts.mutexLock++
}

func (s *AnalysisState) addMutexUnlock() {
	s.counts.mutexUnlock++
}

func (s *AnalysisState) addRWMutexLock() {
	s.counts.rwMutexLock++
}

func (s *AnalysisState) addRWMutexUnlock() {
	s.counts.rwMutexUnlock++
}

func (s *AnalysisState) addLockerLock() {
	s.counts.lockerLock++
}

func (s *AnalysisState) addLockerUnlock() {
	s.counts.lockerUnlock++
}

func (s *AnalysisState) addCondLock() {
	s.counts.condLock++
}

func (s *AnalysisState) addCondUnlock() {
	s.counts.condUnlock++
}

func (s *AnalysisState) addCondWait() {
	s.counts.condWait++
}

func (s *AnalysisState) addCondSignal() {
	s.counts.condSignal++
}

func (s *AnalysisState) addCondBroadcast() {
	s.counts.condBroadcast++
}

func (s *AnalysisState) addCondNew() {
	s.counts.condNew++
}

func (s *AnalysisState) addOnceDo() {
	s.counts.onceDo++
}

func (s *AnalysisState) addUnknownDone() {
	s.counts.unknownDone++
}

func (s *AnalysisState) addUnknownAdd() {
	s.counts.unknownAdd++
}

func (s *AnalysisState) addUnknownWait() {
	s.counts.unknownWait++
}

func (s *AnalysisState) addUnknownLock() {
	s.counts.unknownLock++
}

func (s *AnalysisState) addUnknownUnlock() {
	s.counts.unknownUnlock++
}

func (s *AnalysisState) addUnknownSignal() {
	s.counts.unknownSignal++
}

func (s *AnalysisState) addUnknownBroadcast() {
	s.counts.unknownBroadcast++
}

func (s *AnalysisState) addUnknownDo() {
	s.counts.unknownDo++
}

func splitTarget(target string) string {
	parts := strings.Split(target, ".")
	return parts[len(parts)-1]
}

func targetPieces(target string) int {
	parts := strings.Split(target, ".")
	return len(parts)
}

func stateHeaders() []string {
	res := []string{"fileName", "waitGroupDecls", "condDecls", "onceDecls",
		"mutexDecls", "rwMutexDecls", "lockerDecls", "waitGroupDone",
		"waitGroupAdd", "waitGroupWait", "mutexLock", "mutexUnlock",
		"rwMutexLock", "rwMutexUnlock", "lockerLock", "lockerUnlock",
		"condLock", "condUnlock",
		"condWait", "condSignal", "condBroadcast", "condNew",
		"onceDo", "unknownDone", "unknownAdd", "unknownWait",
		"unknownLock", "unknownUnlock", "unknownSignal", "unknownBroadcast",
		"unknownDo",
	}
	return res
}

func (s *AnalysisState) stateToSlice(fileName string) []string {
	res := []string{fileName, strconv.Itoa(s.counts.waitGroupDecls),
		strconv.Itoa(s.counts.condDecls), strconv.Itoa(s.counts.onceDecls),
		strconv.Itoa(s.counts.mutexDecls), strconv.Itoa(s.counts.rwMutexDecls),
		strconv.Itoa(s.counts.lockerDecls), strconv.Itoa(s.counts.waitGroupDone),
		strconv.Itoa(s.counts.waitGroupAdd), strconv.Itoa(s.counts.waitGroupWait),
		strconv.Itoa(s.counts.mutexLock), strconv.Itoa(s.counts.mutexUnlock),
		strconv.Itoa(s.counts.rwMutexLock), strconv.Itoa(s.counts.rwMutexUnlock),
		strconv.Itoa(s.counts.lockerLock), strconv.Itoa(s.counts.lockerUnlock),
		strconv.Itoa(s.counts.condLock), strconv.Itoa(s.counts.condUnlock),
		strconv.Itoa(s.counts.condWait), strconv.Itoa(s.counts.condSignal),
		strconv.Itoa(s.counts.condBroadcast), strconv.Itoa(s.counts.condNew),
		strconv.Itoa(s.counts.onceDo), strconv.Itoa(s.counts.unknownDone),
		strconv.Itoa(s.counts.unknownAdd), strconv.Itoa(s.counts.unknownWait),
		strconv.Itoa(s.counts.unknownLock), strconv.Itoa(s.counts.unknownUnlock),
		strconv.Itoa(s.counts.unknownSignal), strconv.Itoa(s.counts.unknownBroadcast),
		strconv.Itoa(s.counts.unknownDo),
	}
	return res
}

func (s *AnalysisState) addDone(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == WaitGroup {
				fmt.Printf("Found use of Done for WaitGroup target %s\n", vs[0].name)
				s.addWaitGroupDone()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Done\n", target)
				s.addUnknownDone()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Done\n", target)
			s.addUnknownDone()
		}
	} else {
		fmt.Printf("No match for target %s for call to Done\n", target)
		s.addUnknownDone()
	}
}

func (s *AnalysisState) addAdd(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == WaitGroup {
				fmt.Printf("Found use of Add for WaitGroup target %s\n", vs[0].name)
				s.addWaitGroupAdd()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Add\n", target)
				s.addUnknownAdd()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Add\n", target)
			s.addUnknownAdd()
		}
	} else {
		fmt.Printf("No match for target %s for call to Add\n", target)
		s.addUnknownAdd()
	}
}

func (s *AnalysisState) addWait(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == WaitGroup {
				fmt.Printf("Found use of Wait for WaitGroup target %s\n", vs[0].name)
				s.addWaitGroupWait()
			} else if vs[0].typeof == Cond {
				fmt.Printf("Found use of Wait for Cond target %s\n", vs[0].name)
				s.addCondWait()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Wait\n", target)
				s.addUnknownWait()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Wait\n", target)
			s.addUnknownWait()
		}
	} else {
		fmt.Printf("No match for target %s for call to Wait\n", target)
		s.addUnknownWait()
	}
}

func (s *AnalysisState) addLock(target string) {
	_, ok := s.decls[target]
	if !ok && targetPieces(target) > 1 && splitTarget(target) == "L" {
		target = target[:len(target)-2]
	}
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == Cond {
				fmt.Printf("Found use of Lock for Cond target %s\n", vs[0].name)
				s.addCondLock()
			} else if vs[0].typeof == Mutex {
				fmt.Printf("Found use of Lock for Mutex target %s\n", vs[0].name)
				s.addMutexLock()
			} else if vs[0].typeof == RWMutex {
				fmt.Printf("Found use of Lock for RWMutex target %s\n", vs[0].name)
				s.addRWMutexLock()
			} else if vs[0].typeof == Locker {
				fmt.Printf("Found use of Lock for Locker target %s\n", vs[0].name)
				s.addLockerLock()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Lock\n", target)
				s.addUnknownLock()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Lock\n", target)
			s.addUnknownLock()
		}
	} else {
		fmt.Printf("No match for target %s for call to Lock\n", target)
		s.addUnknownLock()
	}
}

func (s *AnalysisState) addUnlock(target string) {
	_, ok := s.decls[target]
	if !ok && targetPieces(target) > 1 && splitTarget(target) == "L" {
		target = target[:len(target)-2]
	}
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == Cond {
				fmt.Printf("Found use of Unlock for Cond target %s\n", vs[0].name)
				s.addCondUnlock()
			} else if vs[0].typeof == Mutex {
				fmt.Printf("Found use of Unlock for Mutex target %s\n", vs[0].name)
				s.addMutexUnlock()
			} else if vs[0].typeof == RWMutex {
				fmt.Printf("Found use of Unlock for RWMutex target %s\n", vs[0].name)
				s.addRWMutexUnlock()
			} else if vs[0].typeof == Locker {
				fmt.Printf("Found use of Unlock for Locker target %s\n", vs[0].name)
				s.addLockerUnlock()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Unlock\n", target)
				s.addUnknownUnlock()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Unlock\n", target)
			s.addUnknownUnlock()
		}
	} else {
		fmt.Printf("No match for target %s for call to Unlock\n", target)
		s.addUnknownUnlock()
	}
}

func (s *AnalysisState) addSignal(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == Cond {
				fmt.Printf("Found use of Signal for Cond target %s\n", vs[0].name)
				s.addCondSignal()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Signal\n", target)
				s.addUnknownSignal()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Signal\n", target)
			s.addUnknownSignal()
		}
	} else {
		fmt.Printf("No match for target %s for call to Signal\n", target)
		s.addUnknownSignal()
	}
}

func (s *AnalysisState) addBroadcast(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == Cond {
				fmt.Printf("Found use of Broadcast for Cond target %s\n", vs[0].name)
				s.addCondBroadcast()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Broadcast\n", target)
				s.addUnknownBroadcast()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Broadcast\n", target)
			s.addUnknownBroadcast()
		}
	} else {
		fmt.Printf("No match for target %s for call to Broadcast\n", target)
		s.addUnknownBroadcast()
	}
}

func (s *AnalysisState) addDo(target string) {
	target = splitTarget(target)
	vs, ok := s.decls[target]
	if ok {
		if len(vs) == 1 {
			if vs[0].typeof == Once {
				fmt.Printf("Found use of Do for Once target %s\n", vs[0].name)
				s.addOnceDo()
			} else {
				fmt.Printf("Unexpected match for target %s for call to Do\n", target)
				s.addUnknownDo()
			}
		} else {
			fmt.Printf("Multiple matches for target %s for call to Do\n", target)
			s.addUnknownDo()
		}
	} else {
		fmt.Printf("No match for target %s for call to Do\n", target)
		s.addUnknownDo()
	}
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "filePath", "", "The file to be processed")

	var dirPath string
	flag.StringVar(&dirPath, "dirPath", "", "The directory to be processed")

	var outputFile string
	flag.StringVar(&outputFile, "output", "", "The CSV file to be created")

	flag.Parse()

	if outputFile != "" {
		csvFile, err := os.Create(outputFile)
		if err == nil {
			writer := csv.NewWriter(csvFile)
			if err := writer.Write(stateHeaders()); err != nil {
				log.Fatalln("Error writing CSV file", err)
			}
			defer writer.Flush()
			if dirPath != "" {
				fmt.Printf("Processing all go files in directory %s\n", dirPath)
				processDir(dirPath, writer)
			} else if filePath != "" {
				fmt.Printf("Processing file %s\n", filePath)
				processFile(filePath, writer)
			} else {
				fmt.Print("No file or directory given\n")
			}

			csvFile.Close()
		} else {
			log.Fatalln("Error creating CSV file", err)
		}
	} else {
		fmt.Print("No output file given\n")
	}

}

func processDir(dirPath string, writer *csv.Writer) {
	var err = filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Encountered an error accessing path %q: %v\n", path, err)
			return err
		} else {
			if filepath.Ext(path) == ".go" {
				fmt.Printf("Processing file %s\n", path)
				processFile(path, writer)
				return nil
			} else {
				return nil
			}
		}
	})

	if err != nil {
		fmt.Printf("Encountered an error walking the directory tree %q: %v", dirPath, err)
	}
}

func processFile(filePath string, writer *csv.Writer) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		log.Printf("Could not process file %s", filePath)
		log.Print(err)
	} else {
		fileState := &AnalysisState{decls: map[string][]Declaration{}}
		declVisitor := &Visitor{fset: fset, mode: true, state: fileState}
		ast.Walk(declVisitor, file)
		usesVisitor := &Visitor{fset: fset, mode: false, state: fileState}
		ast.Walk(usesVisitor, file)
		if err := writer.Write(fileState.stateToSlice(filePath)); err != nil {
			log.Fatalln("Error writing CSV file", err)
		}
		writer.Flush()
	}
}

type Visitor struct {
	fset  *token.FileSet
	mode  bool
	state *AnalysisState
}

func (v *Visitor) addDef(d Declaration) {
	v.state.addDecl(d)
}

// We should have a consistent return type with information on what
// the match has found
//func matchMakeCall(x *ast.CallExpr, v *Visitor, n ast.Node) {
//	id, ok := x.Fun.(*ast.Ident)
//	if ok {
//		if id.Name == "make" {
//			if len(x.Args) == 1 {
//				t, ok := x.Args[0].(*ast.ChanType)
//				if ok {
//					tname, ok := t.Value.(*ast.Ident)
//					if ok {
//						fmt.Printf("Found a channel of type %s at %s\n", tname.Name, v.fset.Position(n.Pos()))
//					}
//				}
//			} else if len(x.Args) == 2 {
//				t, ok := x.Args[0].(*ast.ChanType)
//				if ok {
//					tname, ok := t.Value.(*ast.Ident)
//					if ok {
//						bsize, ok := x.Args[1].(*ast.BasicLit)
//						if ok {
//							if bsize.Kind == token.INT {
//								fmt.Printf("Found a channel of type %s with literal buffer size %s at %s\n", tname.Name, bsize.Value, v.fset.Position(n.Pos()))
//							} else {
//								fmt.Printf("Found a channel of type %s with buffer size %s at %s\n", tname.Name, bsize.Value, v.fset.Position(n.Pos()))
//							}
//						} else {
//							fmt.Printf("Found a channel of type %s with computed buffer size %s at %s\n", tname.Name, x.Args[1], v.fset.Position(n.Pos()))
//						}
//					}
//				}
//			}
//		}
//	}
//}

//func matchSendStmt(x *ast.SendStmt, v *Visitor, n ast.Node) {
//	id, ok := x.Chan.(*ast.Ident)
//	if ok {
//		fmt.Printf("Found a send to channel %s for value %s\n", id.Name, x.Value)
//	}
//}

//func matchUnaryExpr(x *ast.UnaryExpr, v *Visitor, n ast.Node) {
//	if x.Op == token.ARROW {
//		id, ok := x.X.(*ast.Ident)
//		if ok {
//			fmt.Printf("Found a read of channel %s\n", id.Name)
//		} else {
//			fmt.Printf("Found a read of channel from expr %s\n", x.X)
//		}
//	}
//}

func matchWaitGroupDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "WaitGroup" {
							fmt.Printf("Found declaration of waitgroup %s\n", id.Name)
							v.addDef(createDecl(id.Name, WaitGroup))
							v.state.addWaitGroupDecl()
						}
					}
				}
			}
		}
	}
}

func matchWaitGroupParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "WaitGroup" {
					fmt.Printf("Found declaration of WaitGroup field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, WaitGroup))
					v.state.addWaitGroupDecl()
				}
			}
		}
	}
}

func matchMutexDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Mutex" {
							fmt.Printf("Found declaration of mutex %s\n", id.Name)
							v.addDef(createDecl(id.Name, Mutex))
							v.state.addMutexDecl()
						}
					}
				}
			}
		}
	}
}

func matchMutexParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Mutex" {
					fmt.Printf("Found declaration of Mutex field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, Mutex))
					v.state.addMutexDecl()
				}
			}
		}
	}
}

func matchRWMutexDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "RWMutex" {
							fmt.Printf("Found declaration of rwmutex %s\n", id.Name)
							v.addDef(createDecl(id.Name, RWMutex))
							v.state.addRWMutexDecl()
						}
					}
				}
			}
		}
	}
}

func matchRWMutexParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "RWMutex" {
					fmt.Printf("Found declaration of RWMutex field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, RWMutex))
					v.state.addRWMutexDecl()
				}
			}
		}
	}
}

func matchLockerDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	if len(x.Specs) > 0 {
		spec, ok := x.Specs[0].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Locker" {
							fmt.Printf("Found declaration of locker %s\n", id.Name)
							v.addDef(createDecl(id.Name, Locker))
							v.state.addLockerDecl()
						}
					}
				}
			}
		}
	}
}

func matchLockerParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Locker" {
					fmt.Printf("Found declaration of Locker field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, Locker))
					v.state.addLockerDecl()
				}
			}
		}
	}
}

func matchOnceDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Once" {
							fmt.Printf("Found declaration of once %s\n", id.Name)
							v.addDef(createDecl(id.Name, Once))
							v.state.addOnceDecl()
						}
					}
				}
			}
		}
	}
}

func matchCondDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Cond" {
							fmt.Printf("Found declaration of cond %s\n", id.Name)
							v.addDef(createDecl(id.Name, Cond))
							v.state.addCondDecl()
						}
					}
				}
			}
		}
	}
}

func matchCondAssignDecl(x *ast.AssignStmt, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Rhs); i++ {
		call, ok := x.Rhs[i].(*ast.CallExpr)
		if ok {
			t, ok := call.Fun.(*ast.SelectorExpr)
			if ok {
				tid, ok := t.X.(*ast.Ident)
				if ok {
					if tid.Name == "sync" && t.Sel.Name == "NewCond" {
						id, ok := x.Lhs[0].(*ast.Ident)
						if ok {
							fmt.Printf("Found declaration of cond %s\n", id.Name)
							v.addDef(createDecl(id.Name, Cond))
							v.state.addCondDecl()
						}
					}
				}
			}
		}
	}
}

func matchOnceParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Once" {
					fmt.Printf("Found declaration of Once field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, Once))
					v.state.addOnceDecl()
				}
			}
		}
	}
}

func getFieldType(x *ast.Field) *ast.SelectorExpr {
	fieldType, ok := x.Type.(*ast.SelectorExpr)
	if ok {
		return fieldType
	} else {
		starType, ok := x.Type.(*ast.StarExpr)
		if ok {
			innerFieldType, ok := starType.X.(*ast.SelectorExpr)
			if ok {
				return innerFieldType
			}
		}
	}
	return nil
}

func matchCondParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]

		fieldType := getFieldType(x)

		if fieldType != nil {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Cond" {
					fmt.Printf("Found declaration of cond field %s\n", fieldName.Name)
					v.addDef(createDecl(fieldName.Name, Cond))
					v.state.addCondDecl()
				}
			}
		}
	}
}

func matchDone(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Done" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Done on node %s\n", buf.String())
		v.state.addDone(buf.String())
	}
}

func matchAdd(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Add" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Add on node %s\n", buf.String())
		v.state.addAdd(buf.String())
	}
}

func matchLock(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Lock" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Lock on node %s\n", buf.String())
		v.state.addLock(buf.String())
	}
}

func matchUnlock(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Unlock" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Unlock on node %s\n", buf.String())
		v.state.addUnlock(buf.String())
	}
}

func matchWait(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Wait" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Wait on node %s\n", buf.String())
		v.state.addWait(buf.String())
	}
}

func matchSignal(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Signal" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Signal on node %s\n", buf.String())
		v.state.addSignal(buf.String())
	}
}

func matchBroadcast(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Broadcast" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Broadcast on node %s\n", buf.String())
		v.state.addBroadcast(buf.String())
	}
}

func matchDo(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Do" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Do on node %s\n", buf.String())
		v.state.addDo(buf.String())
	}
}

func matchNewCond(x *ast.CallExpr, v *Visitor, n ast.Node) {
	target, ok := x.Fun.(*ast.SelectorExpr)
	if ok {
		targetName, ok := target.X.(*ast.Ident)
		if ok {
			funName := target.Sel
			if funName.Name == "NewCond" && targetName.Name == "sync" {
				fmt.Print("Found call of NewCond\n")
				v.state.addCondNew()
			}
		}
	}
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	if v.mode {
		if n == nil {
			return nil
		}

		switch x := n.(type) {
		case *ast.GenDecl:
			matchWaitGroupDecl(x, v, n)
			matchMutexDecl(x, v, n)
			matchRWMutexDecl(x, v, n)
			matchLockerDecl(x, v, n)
			matchOnceDecl(x, v, n)
			matchCondDecl(x, v, n)
		case *ast.Field:
			matchWaitGroupParamDecl(x, v, n)
			matchMutexParamDecl(x, v, n)
			matchRWMutexParamDecl(x, v, n)
			matchLockerParamDecl(x, v, n)
			matchOnceParamDecl(x, v, n)
			matchCondParamDecl(x, v, n)
		case *ast.AssignStmt:
			matchCondAssignDecl(x, v, n)
		}
		return v
	} else {
		if n == nil {
			return nil
		}

		switch x := n.(type) {
		case *ast.CallExpr:
			matchNewCond(x, v, n)
		//case *ast.CallExpr:
		//	matchMakeCall(x, v, n)
		//case *ast.SendStmt:
		//	matchSendStmt(x, v, n)
		//case *ast.UnaryExpr:
		//	matchUnaryExpr(x, v, n)
		case *ast.SelectorExpr:
			matchDone(x, v, n)
			matchAdd(x, v, n)
			matchWait(x, v, n)
			matchLock(x, v, n)
			matchUnlock(x, v, n)
			matchSignal(x, v, n)
			matchBroadcast(x, v, n)
			matchDo(x, v, n)
		}
		return v
	}
}
