# Public variables
OUTPUT_DIR ?= out

# Private variables
obj = $(shell ls docs/*.md | sed -r 's@docs/(.*).md@\1@g')
mta = $(wildcard *.md)
changelog_exists = $(wildcard CHANGELOG.txt)
origin_exists = $(wildcard ORIGIN.txt)
is_git_repo = $(wildcard .git)
current_dir = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
formats = pdf slides.pdf html slides.html epub odt txt
all: build

# Build
build: build/archive
$(addprefix build/,$(obj)):
	$(MAKE) build-pdf/$(subst build/,,$@) build-slides.pdf/$(subst build/,,$@) build-html/$(subst build/,,$@) build-slides.html/$(subst build/,,$@) build-epub/$(subst build/,,$@) build-odt/$(subst build/,,$@) build-gmi/$(subst build/,,$@) build-txt/$(subst build/,,$@)

# Build PDF
$(addprefix build-pdf/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --template=./eisvogel-custom.latex --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs -M titlepage=true -M toc=true -M toc-own-page=true -M linkcolor=blue --pdf-engine=xelatex -o "$(OUTPUT_DIR)/$(subst build-pdf/,,$@).pdf" "docs/$(subst build-pdf/,,$@).md"
	
# Build PDF slides
$(addprefix build-slides.pdf/,$(obj)): build/qr
ifeq ($(DISABLE_PDF_SLIDES),true)
	exit 0
else
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to beamer --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs --slide-level=3 --variable theme=metropolis --pdf-engine=xelatex -o "$(OUTPUT_DIR)/$(subst build-slides.pdf/,,$@).slides.pdf" "docs/$(subst build-slides.pdf/,,$@).md"
endif

# Build HTML
$(addprefix build-html/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to markdown --shift-heading-level-by=-1 --resource-path=docs --standalone "docs/$(subst build-html/,,$@).md" | pandoc --to html5 --citeproc --listings --shift-heading-level-by=1 --number-sections --resource-path=docs --toc --katex --self-contained --number-offset=1 -o "$(OUTPUT_DIR)/$(subst build-html/,,$@).html"

# Build HTML slides
$(addprefix build-slides.html/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to slidy --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs --toc --katex --self-contained -o "$(OUTPUT_DIR)/$(subst build-slides.html/,,$@).slides.html" "docs/$(subst build-slides.html/,,$@).md"

# Build EPUB
$(addprefix build-epub/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to epub --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs -M titlepage=true -M toc=true -M toc-own-page=true -M linkcolor=blue -o "$(OUTPUT_DIR)/$(subst build-epub/,,$@).epub" "docs/$(subst build-epub/,,$@).md"

# Build ODT
$(addprefix build-odt/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to odt --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs -M titlepage=true -M toc=true -M toc-own-page=true -M linkcolor=blue -o "$(OUTPUT_DIR)/$(subst build-odt/,,$@).odt" "docs/$(subst build-odt/,,$@).md"

# Build Gemtext
$(addprefix build-gmi/,$(obj)): build/qr
	rm -rf "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi"
	mkdir -p "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi"
	cd "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi" && pandoc --to html --citeproc --extract-media static --metadata title="intermediate" --resource-path="../../docs" "../../docs/$(subst build-gmi/,,$@).md" | pandoc --read html --to gfm-raw_html | md2gemini -a -p -s | sed -e 's@^=> static/static/@=>static/@g' > "$(subst build-gmi/,,$@).gmi" && [ -d static/static ] && mv -f static/static/* static/; rm -rf static/static
	tar -I 'gzip -9' -cvf "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi.gz" -C "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi" .
	rm -rf "$(OUTPUT_DIR)/$(subst build-gmi/,,$@).gmi"

# Build txt
$(addprefix build-txt/,$(obj)): build/qr
	mkdir -p "$(OUTPUT_DIR)"
	pandoc --to plain --citeproc --listings --shift-heading-level-by=-1 --number-sections --resource-path=docs --toc --self-contained -o "$(OUTPUT_DIR)/$(subst build-txt/,,$@).txt" "docs/$(subst build-txt/,,$@).md"

# Build metadata
build/metadata:
	mkdir -p "$(OUTPUT_DIR)"
ifneq ("$(origin_exists)", "")
	cp ORIGIN.txt "$(OUTPUT_DIR)"/ORIGIN.txt
else ifneq ("$(is_git_repo)", "")
	git remote get-url origin > "$(OUTPUT_DIR)"/ORIGIN.txt
else
	echo "file://$(current_dir)" > "$(OUTPUT_DIR)"/ORIGIN.txt
endif
ifneq ("$(changelog_exists)", "")
	cp CHANGELOG.txt "$(OUTPUT_DIR)"/CHANGELOG.txt
else ifneq ("$(is_git_repo)", "")
	git log > "$(OUTPUT_DIR)"/CHANGELOG.txt
else
	touch "$(OUTPUT_DIR)"/CHANGELOG.txt
endif
	cp LICENSE "$(OUTPUT_DIR)"/LICENSE.txt
	$(foreach mt,$(mta),pandoc --to markdown --shift-heading-level-by=-1 --standalone "$(mt)" | pandoc --to html5 --citeproc --listings --shift-heading-level-by=1 --number-sections --resource-path=docs --toc --katex --self-contained --number-offset=1 -o "$(OUTPUT_DIR)/$(subst .md,.html,$(mt))";)

# Build QR code
build/qr:
	mkdir -p docs/static
ifneq ("$(origin_exists)", "")
	qr "$(shell cat ORIGIN.txt)" > docs/static/qr.png
else ifneq ("$(is_git_repo)", "")
	qr "https://$(shell git remote get-url origin | sed -r 's|^.*@(.*):|\1/|g' | sed 's@.*://@@g' | sed 's/.git$$//g')" > docs/static/qr.png
else
	qr "file://$(current_dir)" > docs/static/qr.png
endif

# Build tarball
build/tarball: build/qr build/metadata
	mkdir -p "$(OUTPUT_DIR)"
	tar cvf "$(OUTPUT_DIR)"/source.tar --exclude-from=.gitignore --exclude=.git --exclude="$(OUTPUT_DIR)" .
	tar uvf "$(OUTPUT_DIR)"/source.tar ./Makefile
	tar uvf "$(OUTPUT_DIR)"/source.tar -C "$(OUTPUT_DIR)" ./CHANGELOG.txt ./ORIGIN.txt
	gzip -9 < "$(OUTPUT_DIR)"/source.tar > "$(OUTPUT_DIR)"/source.tar.gz
	rm "$(OUTPUT_DIR)"/source.tar

# Build tree
build/tree: $(addprefix build/,$(obj)) build/tarball
	mkdir -p "$(OUTPUT_DIR)"
ifneq ("$(origin_exists)", "")
	cd "$(OUTPUT_DIR)" && tree -T "$(shell cat ORIGIN.txt | sed -r 's|^.*@(.*):|\1/|g' | sed 's@.*://@@g' | sed 's/.git$$//g')" --du -h -D -H . -I 'index.html|release.tar.gz|release.zip' -o "index.html"
else ifneq ("$(is_git_repo)", "")
	cd "$(OUTPUT_DIR)" && tree -T "$(shell git remote get-url origin | sed -r 's|^.*@(.*):|\1/|g' | sed 's@.*://@@g' | sed 's/.git$$//g')" --du -h -D -H . -I 'index.html|release.tar.gz|release.zip' -o "index.html"
else
	cd "$(OUTPUT_DIR)" && tree -T "$(notdir $(patsubst %/,%,$(current_dir)))" --du -h -D -H . -I 'index.html|release.tar.gz|release.zip' -o "index.html"
endif

# Build archive
build/archive: build/tree
	mkdir -p "$(OUTPUT_DIR)"
	tar -I 'gzip -9' -cvf "$(OUTPUT_DIR)"/release.tar.gz -C "$(OUTPUT_DIR)" --exclude="release.tar.gz" --exclude="release.zip" $(shell ls $(OUTPUT_DIR))
	rm -f "$(OUTPUT_DIR)"/release.zip
	zip -9 -j -x 'release.tar.gz' -x 'release.zip' -FSr "$(OUTPUT_DIR)"/release.zip "$(OUTPUT_DIR)"/*

# Open
$(foreach o,$(obj),$(foreach f,$(formats),open-$(f)/$(o))):
	xdg-open "$(OUTPUT_DIR)/$(notdir $(subst open-,,$@)).$(subst /,,$(dir $(subst open-,,$@)))"

# Develop
dev: build
	while inotifywait -r -e close_write --exclude 'out' .; do $(MAKE); done
$(foreach o,$(obj),$(foreach f,$(formats),dev-$(f)/$(o))):
	$(MAKE) $(subst dev-,build-,$@)
	while inotifywait -r -e close_write --exclude 'out' .; do $(MAKE) $(subst dev-,build-,$@); done

# Clean
clean:
	rm -rf "$(OUTPUT_DIR)" docs/static/qr.png

# Dependencies
depend:
	pip install pillow qrcode md2gemini --break-system-packages