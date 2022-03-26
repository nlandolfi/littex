halmos_text.tex: halmos.gba
	go run gba3.go -in halmos.gba -m tex > halmos_text.tex

halmos_work.pdf: ./halmos_prelude.tex halmos_text.tex halmos_work.tex
	./latexrun halmos_work.tex

oh: halmos_work.pdf
	open halmos_work.pdf

kierkegaard_text.tex: kierkegaard.gba
	go run gba3.go -in kierkegaard.gba -m tex > kierkegaard_text.tex

kierkegaard_work.pdf: halmos_prelude.tex kierkegaard_text.tex kierkegaard_work.tex
	./latexrun kierkegaard_work.tex

ok: kierkegaard_work.pdf
	open kierkegaard_work.pdf

mathgenomics_slides.tex: mathgenomics.gba slides.tmpl
	go run gba3.go -in mathgenomics.gba -m tmpl -tmpl slides.tmpl > mathgenomics_slides.tex

mathgenomics_slides.pdf: mathgenomics_slides.tex mathgenomics_prelude.tex slidemacros.tex
	./latexrun mathgenomics_slides.tex

os: mathgenomics_slides.pdf
	open mathgenomics_slides.pdf

spell:
	aspell -c text.tex
