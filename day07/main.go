package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type FileType int

const (
	FT_REGULAR FileType = iota
	FT_DIR
)

type File struct {
	Name            string
	Type            FileType
	Size            uint64
	Total           uint64
	TotalCalculated bool
	DirContents     map[string]*File
	Parent          *File
	Resolved        bool
}

type FileSystem struct {
	Root *File
	Cwd  *File
}

func MakeFileSystem() (result FileSystem) {
	result.Root = new(File)
	result.Root.Name = "/"
	result.Root.Type = FT_DIR
	result.Root.DirContents = make(map[string]*File)
	result.Cwd = result.Root
	return
}

func (fs *FileSystem) DebugPrint() {
	fs.Root.DebugPrint(0)
}

func (file *File) DebugPrint(indent int) {
	line := strings.Repeat(" ", indent)
	line += "- "
	line += file.Name
	if file.Type == FT_REGULAR {
		line += fmt.Sprintf(" (file, size=%d)", file.Size)
		fmt.Println(line)
	} else {
		if file.Resolved {
			line += " (dir)"
		} else {
			line += " (unresolved dir)"
		}
		fmt.Println(line)
		for _, child := range file.DirContents {
			child.DebugPrint(indent + 2)
		}
	}
}

func (file *File) Find(name string) (result *File) {
	if file.Type != FT_DIR {
		panic("directory expected")
	}
	result, found := file.DirContents[name]
	if !found {
		panic(fmt.Sprintf("file not found: %s", name))
	}
	return
}

func (file *File) addFile(name string) (result *File) {
	result = new(File)
	result.Name = name
	result.Parent = file
	file.DirContents[name] = result
	return
}

func (file *File) AddDir(name string) {
	newFile := file.addFile(name)
	newFile.Type = FT_DIR
	newFile.DirContents = make(map[string]*File)
}

func (file *File) AddRegularFile(name string, size uint64) {
	newFile := file.addFile(name)
	newFile.Size = size
}

func (file *File) GetTotalSize() uint64 {
	if file.Type == FT_REGULAR {
		return file.Size
	}

	if file.TotalCalculated {
		return file.Total
	}

	if !file.Resolved {
		panic(fmt.Sprintf("calculating total size of an unresolved directory: %s", file.Name))
	}

	for _, child := range file.DirContents {
		file.Total += child.GetTotalSize()
	}
	file.TotalCalculated = true
	return file.Total
}

func (file *File) Traverse(cb func(*File)) {
	cb(file)
	if file.Type != FT_DIR {
		return
	}
	for _, child := range file.DirContents {
		child.Traverse(cb)
	}
}

func ParseCdOutput(fs *FileSystem, arg string) {
	if arg == "/" {
		fs.Cwd = fs.Root
	} else if arg == ".." {
		fs.Cwd = fs.Cwd.Parent
		if fs.Cwd == nil {
			panic("cd: no parent for cwd")
		}
	} else {
		fs.Cwd = fs.Cwd.Find(arg)
		if fs.Cwd.Type != FT_DIR {
			panic("cd: directory expected")
		}
	}
}

func ParseLsOutput(fs *FileSystem, cmdOut []string) {
	if fs.Cwd.Resolved {
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
	fs.Cwd.Resolved = true
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
	fs.Root.Traverse(func(file *File) {
		if file.Type == FT_DIR {
			sz := file.GetTotalSize()
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

	usedSpace := fs.Root.GetTotalSize()

	if usedSpace <= maxUsedSpace {
		panic("already enough space?")
	}

	candidate := usedSpace

	fs.Root.Traverse(func(file *File) {
		if file.Type == FT_DIR {
			sz := file.GetTotalSize()
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
