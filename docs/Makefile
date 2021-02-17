diagrams = $(wildcard ./schemas/*.plantuml)
imgs = $(diagrams:.plantuml=.png)

all: $(imgs)

$(imgs): %.png: %.plantuml
	plantuml -png $<
