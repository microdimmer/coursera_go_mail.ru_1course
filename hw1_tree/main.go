package main

import (
	"fmt"
	"io"
	"os"
	"sort"
)

type byFileName []os.FileInfo

func (a byFileName) Len() int           { return len(a) }
func (a byFileName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byFileName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, isPrintFiles bool) error {
	err := dirTreeDeep(out, path, isPrintFiles, []rune{})
	if err != nil {
		return err
	}
	fmt.Fprint(out, "\n")
	return nil
}

func dirTreeDeep(out io.Writer, path string, isPrintFiles bool, deepSl []rune) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()
	filesSlice, err := file.Readdir(0)
	if err != nil {
		return err
	}

	sort.Sort(byFileName(filesSlice)) //need to sort?

	var idxFolder int
	for idx, val := range filesSlice {
		if val.IsDir() {
			if len(deepSl) != 0 || idx != 0 {
				fmt.Fprint(out, "\n")
			}

			for _, val1 := range deepSl {
				if val1 == '│' {
					fmt.Fprint(out, string(val1))
				}
				fmt.Fprint(out, "\t")
			}
			if idx == len(filesSlice)-1 || (idxFolder >= countFolders(&filesSlice)-1 && !isPrintFiles) {
				fmt.Fprintf(out, "└───%s", val.Name())
				deepSl = append(deepSl, ' ') //empty rune?
			} else {
				fmt.Fprintf(out, "├───%s", val.Name())
				deepSl = append(deepSl, '│')
			}

			idxFolder++
			err := dirTreeDeep(out, path+string(os.PathSeparator)+val.Name(), isPrintFiles, deepSl)
			if err != nil {
				panic(err.Error())
			}
			deepSl = deepSl[:len(deepSl)-1]
		} else {
			if !isPrintFiles {
				continue
			}
			if len(deepSl) != 0 || idx != 0 {
				fmt.Fprint(out, "\n")
			}
			for _, val1 := range deepSl {
				if val1 == '│' {
					fmt.Fprint(out, string(val1))
				}
				fmt.Fprint(out, "\t")

			}
			if idx == len(filesSlice)-1 {
				fmt.Fprintf(out, "└───%s", val.Name())
			} else {
				fmt.Fprintf(out, "├───%s", val.Name())
			}

			if val.Size() != 0 {
				fmt.Fprintf(out, " (%vb)", val.Size())
			} else {
				fmt.Fprint(out, " (empty)")
			}
		}
	}
	return err
}

func countFolders(sl *[]os.FileInfo) int {
	var countFolders int
	for _, val := range *sl {
		if val.IsDir() {
			countFolders++
		}
	}
	return countFolders
}
