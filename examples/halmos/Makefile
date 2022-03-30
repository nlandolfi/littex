halmos_text.tex: halmos.lit
	lit -in halmos.lit -m tex > halmos_text.tex

halmos_work.pdf: ./halmos_prelude.tex halmos_text.tex halmos_work.tex
	../latexrun halmos_work.tex

o: halmos_work.pdf
	open halmos_work.pdf
