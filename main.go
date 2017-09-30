package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var (
	addr = flag.String("http", ":8080", "listen on http")
)

func main() {
	flag.Parse()
	var rd io.Reader = os.Stdin
	if flag.Arg(0) != "" {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		rd = file
	}

	data, err := ioutil.ReadAll(rd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	index := NewIndex()

	dir, _ := filepath.Abs(".")
	index.Parse(dir, data)

	fmt.Printf("Listening on %v\n", *addr)
	err = http.ListenAndServe(*addr, &Server{index})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type Server struct {
	Index *Index
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" || r.URL.Path == "/" {
		err := T.Execute(w, map[string]interface{}{
			"StatCount": statCount,
			"Stats":     statSpecs,
			"Files":     server.Index.Files,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return
	}

	if r.URL.Path == "/file" {
		path := r.FormValue("path")
		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No path specified.")
			return
		}

		annotated, err := server.Index.LoadAnnotatedFile(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(annotated)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

var T = template.Must(template.New("").Funcs(template.FuncMap{
	"mul": func(a, b int) int { return a * b },
}).Parse(`
<html>
<body>
	<select id="file" onchange="fileSelected()">
		{{ range .Files }}
		<option value="{{.Path}}">{{.AbsPath}} {{.Stats}}</option>
		{{ end }}
	</select>
	<div id="source">
	</div>

	<style>
	.line {
		position: relative;
		height: 1.2em;
		overflow: hidden;

		--number-width: 3em;
		--info-width: 20em;
		--tags-width: {{mul .StatCount 2}}em;

		contain: strict;
	}
	.line:hover {
		background: #eee;
	}
	
	.line .number {
		position: absolute;
		display: block;
		left: 0; right: 0; top: 0; bottom: 0;
		width: var(--number-width);
	}
	.line .source {
		position: absolute;
		display: block;
		white-space: pre;
		left: var(--number-width);
		right: calc(var(--info-width) + var(--tags-width));
		top: 0; bottom: 0;
		text-overflow: ellipsis;
		overflow: hidden;
	}
	.line .source .tip {
		display: inline-block;
		width: 5px;
		background: #aaa;
	}
	.line .info {
		position: absolute;
		display: block;
		right: var(--tags-width); top: 0; bottom: 0;
		width: var(--info-width);
		text-overflow: ellipsis;
		overflow: hidden;
	}
	.line .tags {
		position: absolute;
		height: 1.2em;
		display: block;
		right: 0;
		width: var(--tags-width);
	}
	.line .tag {
		position: absolute;
		display: block;
		top: 0; bottom: 0;
		width: 2em;
		overflow: hidden;
		border: 1px solid #eee;

		text-align: center;
	}
	.line .tag .good { display: inline-block; width: 0.8em; color: #aaa; }
	.line .tag .bad  { display: inline-block; width: 0.8em; color: #aaa; }
	
	.line .tag .active.good { color: #000; background: #dfd; }
	.line .tag .active.bad  { color: #000; background: #fdd; }

	{{ range $index, $stat := .Stats }}
	.line .tag-{{$index}} { left: {{mul $index 2}}em; }	
	{{ end }}
	</style>

	<script>
		var pending = null;
		function fileSelected() {
			if(pending){
				pending.abort();
			}
			var el = document.getElementById("file")
			if(el.value != ""){
				pending = fetch("/file?path=" + encodeURI(el.value))
					.then(function(response){
						pending = null;
						if(response.ok){
							response.json().then(updateSource);
						}
					})
			}
		}

		function updateSource(file) {
			var fragment = document.createDocumentFragment();
			file.lines.forEach((line, index) => {
				var lineel = h("div", "line");
				lineel.appendChild(h("span", "number", index + 1));

				var source = h("span", "source");
				var p = 0;
				var noteIndex = 0;
				while(noteIndex < line.notes.length){
					var note = line.notes[noteIndex];
					if(note.column < 0){
						noteIndex++;
						continue;
					}
					var text = line.source.substr(p, note.column - p);
					source.appendChild(document.createTextNode(text));
					p = note.column;
					noteIndex++;

					var tip = document.createElement("span");
					tip.className = "tip";
					tip.title = note.message;
					tip.innerText = " ";
					while((noteIndex < line.notes.length) && (line.notes[noteIndex].column == p)){
						tip.title += "\n" + line.notes[noteIndex].message;
						noteIndex++;
					}
					source.appendChild(tip);
				}
				source.appendChild(document.createTextNode(line.source.substr(p)));
				lineel.appendChild(source);
	
				var fullinfo = "";
				if(line.notes.length > 0){
					var infoel = h("span", "info", line.notes[0].message);
					line.notes.forEach(note => {
						fullinfo += note.message + "\n";
					});
					infoel.title = fullinfo;
					lineel.appendChild(infoel);
				}

				var tags = h("span", "tags");
				lineel.appendChild(tags);

				function addtag(i, good, bad){
					var goodCount = 0;
					var badCount = 0;
					
					line.notes.forEach(note => {
						good.forEach(keyword => {
							if(note.message.indexOf(keyword) >= 0){
								goodCount++;
							}
						});
						bad.forEach(keyword => {
							if(note.message.indexOf(keyword) >= 0){
								badCount++;
							}
						});
					})

					if(goodCount + badCount > 0){
						var goodel = h("span", "good", goodCount);
						if(goodCount > 0) goodel.className += " active";
						goodel.title = good.join("\n");
						
						var badel = h("span", "bad", badCount);
						if(badCount > 0) badel.className += " active";
						badel.title = bad.join("\n");

						var el = h("span", "tag active tag-" + i, [
							goodel, "/", badel
						]);
						tags.appendChild(el);
					} else {
						tags.appendChild(h("span", "tag tag-" + i), "");
					}
				}

				{{range $index, $stat := .Stats }}
				addtag({{$index}}, {{$stat.Good}}, {{$stat.Bad}});
				{{end}}

				fragment.appendChild(lineel);
			});

			var source = document.getElementById("source");
			source.innerText = "";
			source.appendChild(fragment);
		}

		function h(tag, className, children){
			var el = document.createElement(tag);
			el.className = className;

			if((typeof children == "string") || (typeof children == "number")){
				children = [children];
			} else if (typeof children == "undefined") {
				children = [];
			}

			for(var i = 0; i < children.length; i++){
				var child = children[i];
				if(typeof child === "string" || typeof child == "number"){
					el.appendChild(document.createTextNode(child));
				} else {
					el.appendChild(child);
				}
			}
			return el;
		}

		fileSelected();
	</script>
</body>
</html>
`))
