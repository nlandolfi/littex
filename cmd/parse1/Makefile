halmos_text.tex: halmos.gba
	go run parse1.go > halmos_text.tex

halmos_work.pdf: ./halmos_prelude.tex halmos_text.tex halmos_work.tex
	./latexrun halmos_work.tex

o: halmos_work.pdf
	open halmos_work.pdf

spell:
	aspell -c text.tex
