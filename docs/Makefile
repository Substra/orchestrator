diagrams = $(wildcard ./schemas/*.plantuml)
imgs = $(diagrams:.plantuml=.png)
mmds = $(shell find ./ -type f -name '*.mmd')
mermaids = $(mmds:.mmd=.svg)

all: $(imgs) $(mermaids)

$(imgs): %.png: %.plantuml
	plantuml -png $<

$(mermaids): %.svg: %.mmd
	npx -p @mermaid-js/mermaid-cli mmdc -i $< -o $@ -b #f7f7f7
