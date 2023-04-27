package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

type Option struct {
	File string `short:"f" long:"file" description:"interface file path" required:"true"`
}

func main() {
	var opt Option
	_, err := flags.Parse(&opt)
	if err != nil {
		return
	}

	file := opt.File

	fs := token.NewFileSet()
	parsedFile, err := decorator.ParseFile(fs, file, nil, 0)
	if err != nil {
		log.Fatalf("parsing package: %s: %s\n", file, err)
	}

	dirPath := filepath.Dir(file)
	fileName := filepath.Base(file)
	fName := fileName[:len(fileName)-len(path.Ext(fileName))]
	var packageName string
	var structName string
	//var imports []*dst.ImportSpec

	dst.Inspect(parsedFile, func(n dst.Node) bool {
		file, ok := n.(*dst.File)
		if ok {
			packageName = file.Name.Name
			//imports = file.Imports
			structName = getStructName(file)
			if structName == "" {
				panic("未找到 struct")
			}
		}

		decl, ok := n.(*dst.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}

		for _, spec := range decl.Specs {
			typeSpec, _ok := spec.(*dst.TypeSpec)
			if !_ok {
				continue
			}

			var interfaceType *dst.InterfaceType
			if interfaceType, ok = typeSpec.Type.(*dst.InterfaceType); !ok {
				continue
			}

			for _, v := range interfaceType.Methods.List {
				if len(v.Names) > 0 {
					funcName := v.Names[0].String()
					if funcName == "i" {
						continue
					}
					filename := fmt.Sprintf("%s/%s_%s.go", dirPath, fName, strings.ToLower(funcName))
					funcDecl := &dst.FuncDecl{}
					funcDecl.Recv = &dst.FieldList{
						Opening: true,
						Closing: true,
						List: []*dst.Field{
							{
								Names: []*dst.Ident{
									dst.NewIdent("m"),
								},
								Type: &dst.StarExpr{
									X: dst.NewIdent(structName),
								},
							},
						},
					}
					funcDecl.Name = dst.NewIdent(funcName)
					if t, ok := v.Type.(*dst.FuncType); ok {
						funcDecl.Type = t
					}
					funcDecl.Body = &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun: dst.NewIdent("panic"),
									Args: []dst.Expr{
										&dst.BasicLit{
											Kind:  token.STRING,
											Value: "\"implement me\"",
										},
									},
								},
								Decs: dst.ExprStmtDecorations{
									NodeDecs: dst.NodeDecs{
										Before: dst.NewLine,
										Start: []string{
											"// TODO: implement me",
										},
										After: dst.NewLine,
									},
								},
							},
						},
					}

					dstFile := &dst.File{
						Name: dst.NewIdent(packageName),
						Imports: []*dst.ImportSpec{
							{
								Path: &dst.BasicLit{
									Kind:  token.STRING,
									Value: "\"gorm.io/gorm\"",
								},
							},
						},
						Decls: []dst.Decl{
							funcDecl,
						},
					}

					err = FmtPrint(filename, dstFile)
					if err != nil {
						continue
					}
				}
			}
		}
		return true
	})
}

func FmtPrint(filename string, dstFile *dst.File) (err error) {
	fmt.Println("  └── file : ", filename)
	buf := bytes.NewBuffer([]byte{})
	err = decorator.Fprint(buf, dstFile)
	if err != nil {
		fmt.Println("输出代码时出错：", err)
		return
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("格式化代码时出错：", err)
		return
	}
	outputFile, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_TRUNC|os.O_RDWR, 0o766)
	if err != nil {
		fmt.Println("创建文件时出错:", err)
		return
	}
	if outputFile == nil {
		fmt.Printf("output file is nil \n")
		return
	}
	defer outputFile.Close()
	_, err = outputFile.Write(formatted)
	if err != nil {
		fmt.Println("写入文件时出错：", err)
		return
	}
	return
}

func getStructName(file *dst.File) string {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*dst.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*dst.StructType); ok {
						return typeSpec.Name.Name
					}
				}
			}
		}
	}
	return ""
}
