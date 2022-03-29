package main

import (
	//"encoding/json"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	//"io"

	"log"
	"net/http"
	"os"
)

//------------ Переменные ----------------------------------------------------
var (
	err           error
	post_template = template.Must(template.ParseFiles(
		"html/main.html",
		"html/templates.html"))
	settings *settings_t
)

//------------ Структуры ----------------------------------------------------
type ( //settings_t, db_t, site_t
	settings_site_t struct {
		IP   string
		Port string
	}
	settings_src_t struct {
		Name string
		Dir  string
		Ext  []string
	}
	settings_t struct {
		Site settings_site_t
		Src  []settings_src_t
	}
)
type tree_t struct {
	Path     string    `json:"-"`
	Name     string    `json:"text"`
	IsDir    bool      `json:"-"`
	State    string    `json:"state"`
	Icon     string    `json:"icon"`
	Children *[]tree_t `json:"children"`
}

//------------ HTTP функции ----------------------------------------------------
func page_main(w http.ResponseWriter, r *http.Request) {
	err = post_template.ExecuteTemplate(w, "main", nil)
	if err != nil {
		fmt.Printf("Error: func page_main: post_template.ExecuteTemplate(): %s\n", err)
		http.Error(w, http.StatusText(500), 500)
	}
}
func page_GetDirs(w http.ResponseWriter, r *http.Request) {
	dirs := new([]tree_t)
	for _, v := range settings.Src { //перебираем все источники из settings
		var d tree_t
		d.Name = filepath.Base(v.Name)
		d.Path = v.Dir
		d.IsDir = true
		d.State = "{\"opened\": false}"
		d.Icon = ""
		d.Children = DirList(filepath.Join(v.Dir)) // поиск в папке
		*dirs = append(*dirs, d)
	}
	js, _ := json.Marshal(dirs)
	_, err := w.Write(js)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
func page_GetHLStyles(w http.ResponseWriter, r *http.Request) {
	//json, _ := json.Marshal(dirs)
	var css []string
	var path = filepath.Join("lib", "highlight", "styles")
	filepath.Walk(path, func(p string, info fs.FileInfo, err error) error { // перебор файлов в каталоге
		if !info.IsDir() && filepath.Ext(p) == ".css" { // берём только файлы css
			css = append(css, p)
		}
		return nil
	})
	json, _ := json.Marshal(css)
	_, err = w.Write(json)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
func page_GetContent(w http.ResponseWriter, r *http.Request) {
	type file_t struct {
		Name string
		Body string
	}
	type vars_t struct {
		Path  string
		Files []file_t
	}
	var (
		vars vars_t
		p    []string // path array
		path = r.FormValue("path")
	)
	path = strings.ReplaceAll(path, "/", string(os.PathSeparator)) // меняем разделители в строке на те, что в ОС
	p = strings.Split(path, string(os.PathSeparator))              // разделяем путь по разделителю текущей ОС
	for _, v := range settings.Src {                               // вместо псевдонима подставляем полный путь
		if v.Name == p[0] {
			path = filepath.Join(v.Dir, filepath.Join(p[1:]...)) // объединяем пути
			break
		}
	}
	filepath.Walk(path, func(p string, info fs.FileInfo, err error) error { // перебор файлов в каталоге
		if (path == p) && !info.IsDir() { // берём только путь равный исходному и проверяем IsDir
			path = filepath.Dir(p)
		}
		return nil
	})
	for _, s := range settings.Src {
		if strings.Contains(path, s.Dir) {
			vars.Path = strings.ReplaceAll(path, s.Dir, s.Name) //заменяем путь к каталогу на псевдоним
		} else {
			continue
		}
		for _, e := range s.Ext { //перебор списка расширений
			d, err := ioutil.ReadDir(path) //чтение всех файлов и директорий в path
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, f := range d {
				if (e == filepath.Ext(f.Name())) && !f.IsDir() { // если совпадает расширение и неПапка
					data, err := ioutil.ReadFile(filepath.Join(path, f.Name())) // чтение файла
					if err != nil {
						fmt.Println(err)
						continue
					}
					vars.Files = append(vars.Files, file_t{f.Name(), string(data)}) // добавляем
				}
			}
		}
	}
	if len(vars.Files) == 0 {
		vars.Files = append(vars.Files, file_t{"NO FILES", ""})
	}
	json, _ := json.Marshal(vars)
	_, err = w.Write(json)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

//------------ Общие функции ---------------------------------------------------
func settings_init(f string) *settings_t {
	file, _ := os.Open(f)
	decoder := json.NewDecoder(file)
	config := new(settings_t)
	decoder.Decode(config)
	return config
}
func DirList(path string) (dirs *[]tree_t) { // рекурсивно просматривает папки, составляет <ul> список
	dirs = new([]tree_t)
	lst, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, v := range lst { // сначала в списке ищем папки и добавляем в dirs
		if v.IsDir() {
			var d tree_t
			d.Name = v.Name()
			d.Path = filepath.Join(path, v.Name())
			d.IsDir = v.IsDir()
			d.State = "{\"opened\": false}"
			d.Icon = ""
			d.Children = DirList(filepath.Join(path, v.Name()))
			*dirs = append(*dirs, d)
		}
	}
	for _, s := range settings.Src { // затем файлы по перечню расширений
		if !strings.Contains(path, s.Dir) { //ищем наш источник в settings
			continue // пропускаем лишнее
		}
		for _, e := range s.Ext {
			for _, v := range lst {
				if (filepath.Ext(v.Name()) == e) && !v.IsDir() {
					var d tree_t
					d.Name = v.Name()
					d.Path = filepath.Join(path, v.Name())
					d.IsDir = v.IsDir()
					d.State = "{\"opened\": false}"
					d.Icon = "jstree-file"
					*dirs = append(*dirs, d)
				}
			}
		}
	}
	return dirs
}

//------------ main ------------------------------------------------------------
func main() {
	settings = settings_init("settings.json") //парсим конфиг файл
	http.HandleFunc("/", page_main)
	http.HandleFunc("/info", page_GetDirs)
	http.HandleFunc("/getcontent", page_GetContent)
	http.HandleFunc("/gethlstyles", page_GetHLStyles)

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("./lib"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))
	http.Handle("/SetImage/", http.StripPrefix("/SetImage/", http.FileServer(http.Dir("./SetImage"))))

	fmt.Printf("WebServer is listening, port: %s\n", settings.Site.Port)
	http.ListenAndServeTLS(fmt.Sprintf(":%s", settings.Site.Port), "./ssl/domain.crt", "./ssl/private.key", nil)
}
