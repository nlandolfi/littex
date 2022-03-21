halmos_text.tex: halmos.gba
	go run parse1.go -in halmos.gba -m tex > halmos_text.tex

halmos_work.pdf: ./halmos_prelude.tex halmos_text.tex halmos_work.tex
	./latexrun halmos_work.tex

# o: halmos_work.pdf
# 	open halmos_work.pdf

kierkegaard_text.tex: kierkegaard.gba
	go run parse1.go -in kierkegaard.gba -m tex > kierkegaard_text.tex

kierkegaard_work.pdf: halmos_prelude.tex kierkegaard_text.tex kierkegaard_work.tex
	./latexrun kierkegaard_work.tex

o: kierkegaard_work.pdf
	open kierkegaard_work.pdf

spell:
	aspell -c text.tex
