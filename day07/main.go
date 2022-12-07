package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type File interface {
	Parent() File
	Name() string
	Size() uint64
}

type Directory struct {
	parent          *Directory
	name            string
	total           uint64
	totalCalculated bool
	resolved        bool
	contents        map[string]File
}

type RegularFile struct {
	parent *Directory
	name   string
	size   uint64
}

func (f *RegularFile) Parent() File {
	return f.parent
}

func (f *RegularFile) Name() string {
	return f.name
}

func (f *RegularFile) Size() uint64 {
	return f.size
}

func (d *Directory) Parent() File {
	return d.parent
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Resolved() bool {
	return d.resolved
}

func (d *Directory) MarkResolved() {
	d.resolved = true
}

func (d *Directory) TraverseContents(cb func(File)) {
	if !d.resolved {
		panic(fmt.Sprintf("traversing an unresolved directory: %s", d.Name))
	}

	for _, child := range d.contents {
		cb(child)
	}
}

func (d *Directory) Size() uint64 {
	if d.totalCalculated {
		return d.total
	}

	if !d.resolved {
		panic(fmt.Sprintf("calculating total size of an unresolved directory: %s", d.Name))
	}

	d.TraverseContents(func(child File) {
		d.total += child.Size()
	})
	d.totalCalculated = true
	return d.total
}

func (d *Directory) Find(name string) (result File) {
	result, found := d.contents[name]
	if !found {
		panic(fmt.Sprintf("file not found: %s", name))
	}
	return
}

func NewDir(name string) *Directory {
	newd := new(Directory)
	newd.name = name
	newd.contents = make(map[string]File)
	return newd
}

func (d *Directory) AddDir(name string) {
	newd := NewDir(name)
	newd.parent = d
	_, exists := d.contents[name]
	if exists {
		panic("adding directory that already exists")
	}
	d.contents[name] = newd
}

func (d *Directory) AddRegularFile(name string, size uint64) {
	newf := new(RegularFile)
	newf.name = name
	newf.parent = d
	newf.size = size

	d.contents[name] = newf
}

type FileSystem struct {
	Root *Directory
	Cwd  *Directory
}

func MakeFileSystem() (result FileSystem) {
	result.Root = NewDir("/")
	result.Cwd = result.Root
	return
}

func debugPrintInner(file File, indent int) {
	line := strings.Repeat(" ", indent)
	line += "- "
	line += file.Name()
	switch t := file.(type) {
	case *RegularFile:
		line += fmt.Sprintf(" (file, size=%d)", file.Size())
	case *Directory:
		if t.Resolved() {
			line += " (dir)"
		} else {
			line += " (unresolved dir)"
		}
		fmt.Println(line)
		t.TraverseContents(func(child File) {
			debugPrintInner(child, indent+2)
		})
	}
}

func (fs *FileSystem) DebugPrint() {
	debugPrintInner(fs.Root, 0)
}

func TraverseDeep(file File, cb func(File)) {
	cb(file)
	dir, ok := file.(*Directory)
	if !ok {
		return
	}
	dir.TraverseContents(func(child File) {
		TraverseDeep(child, cb)
	})
}

func ParseCdOutput(fs *FileSystem, arg string) {
	if arg == "/" {
		fs.Cwd = fs.Root
	} else if arg == ".." {
		var ok bool
		fs.Cwd, ok = fs.Cwd.Parent().(*Directory)
		if !ok {
			panic("cd: parent is not a directory")
		}
		if fs.Cwd == nil {
			panic("cd: no parent for cwd")
		}
	} else {
		var ok bool
		fs.Cwd, ok = fs.Cwd.Find(arg).(*Directory)
		if !ok {
			panic("cd: directory expected")
		}
	}
}

func ParseLsOutput(fs *FileSystem, cmdOut []string) {
	if fs.Cwd.Resolved() {
		panic("ls: consistency check not implemented")
	}

	for _, line := range cmdOut {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			panic("ls: expected 2 fields in output")
		}

		name := fields[1]
		if fields[0] == "dir" {
			fs.Cwd.AddDir(name)
		} else {
			size, err := strconv.ParseUint(fields[0], 10, 64)
			if err != nil {
				panic(err)
			}
			fs.Cwd.AddRegularFile(name, size)
		}
	}
	fs.Cwd.MarkResolved()
}

func ParseCommandOutput(fs *FileSystem, command string, cmdOut []string) {
	if !strings.HasPrefix(command, "$") {
		panic("command line must start with $")
	}

	command = strings.TrimSpace(strings.TrimPrefix(command, "$"))
	fields := strings.Fields(command)
	cmdName := fields[0]
	cmdArgs := fields[1:]

	if cmdName == "cd" {
		if len(cmdArgs) != 1 {
			panic("cmd: Expected exactly 1 argument")
		}
		if len(cmdOut) != 0 {
			panic("cmd: Unexpected output")
		}
		ParseCdOutput(fs, cmdArgs[0])
	} else if cmdName == "ls" {
		if len(cmdArgs) != 0 {
			panic("ls: Unexpected arguments")
		}
		ParseLsOutput(fs, cmdOut)
	}
}

func RunMode1(fs *FileSystem) {
	var sum uint64
	TraverseDeep(fs.Root, func(file File) {
		d, isdir := file.(*Directory)
		if isdir {
			sz := d.Size()
			if sz <= 100000 {
				sum += sz
			}
		}
	})

	fmt.Println(sum)
}

func RunMode2(fs *FileSystem) {
	const totalSpace uint64 = 70000000
	const reqUnusedSpace uint64 = 30000000
	const maxUsedSpace uint64 = totalSpace - reqUnusedSpace

	usedSpace := fs.Root.Size()

	if usedSpace <= maxUsedSpace {
		panic("already enough space?")
	}

	candidate := usedSpace

	TraverseDeep(fs.Root, func(file File) {
		d, isdir := file.(*Directory)
		if isdir {
			sz := d.Size()
			if usedSpace-sz <= maxUsedSpace && sz < candidate {
				candidate = sz
			}
		}
	})

	fmt.Println(candidate)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	fs := MakeFileSystem()

	command := scanner.Text()
	cmdOut := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line[0] == '$' {
			ParseCommandOutput(&fs, command, cmdOut)
			command = line
			cmdOut = cmdOut[:0]
		} else {
			cmdOut = append(cmdOut, line)
		}
	}

	ParseCommandOutput(&fs, command, cmdOut)

	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		RunMode2(&fs)
	} else {
		RunMode1(&fs)
	}
}
