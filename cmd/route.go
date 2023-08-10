package cmd

import (
	"fmt"
	"github.com/farseer-go/fsctl/parse"
	"github.com/farseer-go/fsctl/utils"
	"github.com/farseer-go/utils/file"
	"go/ast"
	"os"
	"strings"
)

func init() {
	c := route{}
	Commands[c.ShortCommand()] = c
	Commands[c.FullCommand()] = c
}

type route struct {
}

func (receiver route) Execute(args []string) {
	rootPath := "../demo/shopping/"
	routePath := rootPath + "route.go"

	receiver.checkRoute(routePath)

	var routeComments []parse.RouteComment
	// 解析整个项目
	parse.ASTDir(rootPath, func(filePath string, astFile *ast.File, funcDecl *ast.FuncDecl) {
		if funcDecl.Doc == nil {
			return
		}
		rc := parse.RouteComment{IocNames: make(map[string]string)}

		// 解析头部注解：区域
		if astFile.Doc != nil {
			for _, comment := range astFile.Doc.List {
				// 去除前缀：// 空格
				commentText := clearCommentPrefix(comment.Text)
				// 解析
				rc.ParsePackageComment(commentText)
			}
		}
		// 解析是否有注解
		for _, comment := range funcDecl.Doc.List {
			// 去除前缀：// 空格
			commentText := clearCommentPrefix(comment.Text)
			// 解析
			rc.ParseFuncComment(commentText)
		}

		// 解析成功
		if rc.IsHaveComment() {
			// 移除相对路径和文件名，得到包路径
			rc.PackagePath = filePath[len(rootPath):strings.LastIndex(filePath, "/")]
			rc.PackagePath = parse.GetRootPackage(rootPath) + "/" + rc.PackagePath
			// 解析函数类型
			rc.ParseFuncType(astFile, funcDecl)
			// 解析Url
			rc.Url = strings.ToLower(strings.Replace(rc.Url, "{area}", rc.Area, -1))
			rc.Url = strings.ToLower(strings.Replace(rc.Url, "{action}", rc.FuncName, -1))
			if !strings.HasPrefix(rc.Url, "/") {
				rc.Url = "/" + rc.Url
			}
			routeComments = append(routeComments, rc)
		}
	})

	// 生成route.go文件
	routeContent := parse.BuildRoute(routePath, routeComments)
	file.WriteString(routePath, routeContent)
}

// 检查根目录route.go文件是否为fsctl工具生成
func (receiver route) checkRoute(routePath string) {
	if file.IsExists(routePath) && !parse.CheckIsRoute(routePath) {
		fmt.Printf(utils.Red("route.go文件不是fsctl工具生成，请手动删除./route.go后再重新运行此命令"))
		os.Exit(0)
	}
	// 删除route.go文件
	file.Delete(routePath)
}

func (receiver route) FullCommand() string {
	return "route"
}

func (receiver route) ShortCommand() string {
	return "-r"
}

func (receiver route) CommandDesc() string {
	return "动态路由配置"
}

// 只取带 // @xxx的注解
func clearCommentPrefix(comment string) string {
	// 去除前缀：//
	comment = strings.TrimPrefix(comment, parse.CommentPrefix)
	// 去除空格
	comment = strings.TrimPrefix(comment, " ")
	// 约定前缀为@标记
	if strings.HasPrefix(comment, "@") {
		return strings.TrimPrefix(comment, "@")
	}
	return ""
}
