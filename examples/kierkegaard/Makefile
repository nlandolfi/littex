kierkegaard_text.tex: kierkegaard.lit
	go run lit.go -in kierkegaard.lit -m tex > kierkegaard_text.tex

kierkegaard_work.pdf: halmos_prelude.tex kierkegaard_text.tex kierkegaard_work.tex
	../latexrun kierkegaard_work.tex

o: kierkegaard_work.pdf
	open kierkegaard_work.pdf
