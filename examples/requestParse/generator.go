package main

/*

parse and generate function ti parse http.Request form

structs may be something like this

type reqProfile struct {
	ID       int64  `req:"id,nempty,nzero",json:"id"`
	Name     string `req:"name,sphinx,nempty,len>3,len<64",json:"name"`
	Language string `req:"language,nempty,len=2",json:"language"`
	Params   map[string]string `req:"prefix=param_",json:"params"`
	Groups   map[string][]string `req:"prefix=group_",json:"groups"`
	Passwd   string `req:"password,len>8",json:"-"`
}

will result in function
func reqProfileParse(request *http.Request) (reqProfile,string) {...}

*/

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kanocz/goparse"
)

func main() {
	s, err := goparse.GetFileStructs(os.Args[1], "req", "req")
	if nil != err {
		fmt.Println("Error parsing file:", err)
		return
	}

	fmt.Print("package main\n\n")
	fmt.Print("import (\n\t\"net/http\"\n\t\"github.com/julienschmidt/httprouter\"\n\t\"strconv\"\nt\"strings\"\n)\n\n")

	for _, st := range s {
		fmt.Printf("func %sParse(request *http.Request, params httprouter.Params) (%s,string) {\n", st.Name, st.Name)
		fmt.Print("\tvar err error\n\t_ = err\n")
		fmt.Printf("\tres := %s{}\n", st.Name)
		fmt.Print("\trequest.ParseMultipartForm(0)\n\n")
		for _, f := range st.Field {

			pname := f.Name
			if nil != f.Tags && len(f.Tags) > 0 {
				pname = f.Tags[0]
			}

			checks := []string{}
			if nil != f.Tags && len(f.Tags) > 1 {
				checks = f.Tags[1:]
			}

			var isParam = false
			if len(checks) > 1 && "param" == checks[0] {
				isParam = true
				checks = checks[1:]
			}

			if f.Type == "map[string][]string" {
				prefix, ok := f.TagParams["prefix"]
				if !ok {
					log.Fatalln("no prefix for map[string][]string field", f.Name, "in struct", st.Name)
				}
				prefixLen := len(prefix)
				fmt.Printf("\tres.%s = map[string][]string{}\n\n", f.Name)
				fmt.Printf("\tfor k, v := range request.Form {\n\t\tif strings.HasPrefix(k, \"%s\") {\n\t\t\tres.%s[k[%d:]] = v\n\t\t}\n\t}\n", prefix, f.Name, prefixLen)
			} else if f.Type == "map[string]string" {
				prefix, ok := f.TagParams["prefix"]
				if !ok {
					log.Fatalln("no prefix for map[string]string field", f.Name, "in struct", st.Name)
				}
				prefixLen := len(prefix)
				fmt.Printf("\tres.%s = map[string]string{}\n\n", f.Name)
				fmt.Printf("\tfor k, v := range request.Form {\n\t\tif strings.HasPrefix(k, \"%s\") {\n\t\t\tres.%s[k[%d:]] = v[0]\n\t\t}\n\t}\n", prefix, f.Name, prefixLen)

			} else if f.Type == "[]int64" {
				fmt.Printf("\tres.%s = []int64{}\n\n", f.Name)
				_, ok := f.TagParams["jarray"]
				if ok {
					if isParam {
						fmt.Printf("\tparam%s := strings.Split(params.ByName(\"%s\"), \",\")\n", f.Name, pname)
					} else {
						fmt.Printf("\tparam%s := strings.Split(request.Form.Get(\"%s\"), \",\")\n", f.Name, pname)
					}
					fmt.Printf("\tfor _, _x := range param%s {\n", f.Name)
					fmt.Print("\t\tx, err := strconv.ParseInt(_x, 10, 64)\n")
					fmt.Printf("\t\tif nil == err {\n")
					fmt.Printf("\t\t\tres.%s = append(res.%s, x)\n", f.Name, f.Name)
					fmt.Print("\t\t}\n")
					fmt.Print("\t}\n")

					// pre convert chcecks
					for _, c := range checks {
						switch c {
						case "nempty":
							fmt.Printf("\tif len(res.%s) == 0 {\n\t\treturn %s{}, \"%s_empty\"\n\t}\n", f.Name, st.Name, pname)
						}
					}

				} else {

				}

				if !ok {
					log.Fatalln("no prefix for map[string]string field", f.Name, "in struct", st.Name)
				}

			} else {

				if !strings.HasPrefix(f.Type, "[]") {
					if isParam {
						fmt.Printf("\tparam%s := params.ByName(\"%s\")\n", f.Name, pname)
					} else {
						fmt.Printf("\tparam%s := request.Form.Get(\"%s\")\n", f.Name, pname)
					}
				} else {
					// todo: other [] types
					if f.Type == "[]string" {
						fmt.Printf("\tparam%s, ok%s := request.Form.[\"%s\"]\n", f.Name, f.Name, pname)
					}
				}

				// pre convert chcecks
				for _, c := range checks {
					switch c {
					case "nempty":
						fmt.Printf("\tif \"\" == param%s {\n\t\treturn %s{}, \"%s_empty\"\n\t}\n", f.Name, st.Name, pname)
					case "sphinx":
						fmt.Printf("\tif !sphinxCheckString(param%s) {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", f.Name, st.Name, pname)
					}
				}

				if tagLen, ok := f.TagParams["len"]; ok {
					fmt.Printf("\tif %s != len(param%s) {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", tagLen, f.Name, st.Name, pname)
				}

				if tagLen, ok := f.TagGt["len"]; ok {
					fmt.Printf("\tif len(param%s) < %d {\n\t\treturn %s{}, \"%s_short\"\n\t}\n", f.Name, tagLen, st.Name, pname)
				}

				if tagLen, ok := f.TagLt["len"]; ok {
					fmt.Printf("\tif len(param%s) > %d {\n\t\treturn %s{}, \"%s_long\"\n\t}\n", f.Name, tagLen, st.Name, pname)
				}

				switch f.Type {
				case "string":
					fmt.Printf("\tres.%s = param%s\n\n", f.Name, f.Name)
				case "[]string":
					fmt.Printf("\tres.%s = param%s\n\n", f.Name, f.Name)
				case "int":
					fmt.Printf("\tres.%s = strconv.Atoi(param%s)\n\n", f.Name, f.Name)
				case "int64":
					fmt.Printf("\tres.%s, err = strconv.ParseInt(param%s, 10, 64)\n\n", f.Name, f.Name)
				case "uint64":
					fmt.Printf("\tres.%s, err = strconv.ParseUint(param%s, 10, 64)\n\n", f.Name, f.Name)
				case "float64":
					fmt.Printf("\tres.%s, err = strconv.ParseFloat(param%s, 64)\n\n", f.Name, f.Name)
				case "bool":
					fmt.Printf("\tres.%s, err = strconv.ParseBool(param%s)\n\n", f.Name, f.Name)
				}

				// post convert chcecks
				for _, c := range checks {
					switch c {
					case "valid":
						fmt.Printf("\tif nil != err {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", st.Name, pname)
					case "nzero":
						fmt.Printf("\tif 0 == res.%s {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", f.Name, st.Name, pname)
					}
				}

				if tagVal, ok := f.TagParams["val"]; ok {
					fmt.Printf("\tif %s != res.%s {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", tagVal, f.Name, st.Name, pname)
				}

				if tagVal, ok := f.TagGt["val"]; ok {
					fmt.Printf("\tif %d > res.%s {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", tagVal, f.Name, st.Name, pname)
				}

				if tagVal, ok := f.TagLt["val"]; ok {
					fmt.Printf("\tif %d < res.%s {\n\t\treturn %s{}, \"%s_invalid\"\n\t}\n", tagVal, f.Name, st.Name, pname)
				}

			}

		}
		fmt.Println("return res, \"\"")
		fmt.Println("}")
	}
}
