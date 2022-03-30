mathgenomics_slides.tex: mathgenomics.lit slides.tmpl
	lit -in mathgenomics.lit -o tmpl -tmpl slides.tmpl > mathgenomics_slides.tex

mathgenomics_slides.pdf: mathgenomics_slides.tex mathgenomics_prelude.tex slidemacros.tex
	../latexrun mathgenomics_slides.tex

o: mathgenomics_slides.pdf
	open mathgenomics_slides.pdf
