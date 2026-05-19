package model

type OutputFormat string

const (
	FormatMarkdown OutputFormat = "markdown"
	FormatHTML     OutputFormat = "html"
	FormatRawHTML  OutputFormat = "raw_html"
	FormatJSON     OutputFormat = "json"
	FormatLinks    OutputFormat = "links"
	FormatMetadata OutputFormat = "metadata"
)

func (f OutputFormat) Valid() bool {
	switch f {
	case FormatMarkdown, FormatHTML, FormatRawHTML, FormatJSON, FormatLinks, FormatMetadata:
		return true
	}
	return false
}

type PDFParseMode string

const (
	PDFModeFast PDFParseMode = "fast"
	PDFModeAuto PDFParseMode = "auto"
	PDFModeOCR  PDFParseMode = "ocr"
)

func (m PDFParseMode) Valid() bool {
	switch m {
	case PDFModeFast, PDFModeAuto, PDFModeOCR:
		return true
	}
	return false
}

type PDFOutput string

const (
	PDFOutputText     PDFOutput = "text"
	PDFOutputMarkdown PDFOutput = "markdown"
	PDFOutputRaw      PDFOutput = "raw"
)

func (o PDFOutput) Valid() bool {
	switch o {
	case PDFOutputText, PDFOutputMarkdown, PDFOutputRaw:
		return true
	}
	return false
}
