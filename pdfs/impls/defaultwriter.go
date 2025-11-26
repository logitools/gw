package impls

import (
	"bytes"
	"io"
	"os"

	"github.com/logitools/gw/pdfs"
	"github.com/logitools/gw/rw"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

// DefaultWriter - simple PDF writer
// Supported unit: "pt" only
// Currently implemented using gofpdf. -> Will be implemented directly
// Currently Custom Size Not Supported: "Letter" and "A4" Only
// ToDo: Support custom fonts
type DefaultWriter struct {
	pdfs.Writer[int] // [Embedded Interface] [To Implement]

	paperSize   pdfs.PaperSize
	orientation string

	impl      *gofpdf.Fpdf
	templates *pdfs.TemplateStore[int]
}

func NewDefaultWriter(paperSize pdfs.PaperSize, orientation string) *DefaultWriter {
	return &DefaultWriter{
		orientation: orientation,
		paperSize:   paperSize,
		impl:        gofpdf.New(orientation, "pt", paperSize.Name, ""),
		templates:   pdfs.NewTemplateStore[int](),
	}
}

func (w *DefaultWriter) PaperSize() pdfs.PaperSize {
	return w.paperSize
}

func (w *DefaultWriter) Orientation() string {
	return w.orientation
}

func (w *DefaultWriter) TemplateStore() *pdfs.TemplateStore[int] {
	return w.templates
}

func (w *DefaultWriter) ImportPageAsTemplate(filepath string, pageNum int, storeKey string) error {
	// Check filepath exist
	tplID := gofpdi.ImportPage(w.impl, filepath, pageNum, "/MediaBox")
	w.templates.Store(storeKey, tplID)
	return nil
}

func (w *DefaultWriter) AddBlankPage() {
	w.impl.AddPage()
}

func (w *DefaultWriter) AddTemplatePage(storeKey string) bool {
	template, ok := w.templates.Get(storeKey)
	if !ok {
		return false
	}
	w.impl.AddPage()
	gofpdi.UseImportedTemplate(w.impl, template, 0, 0, w.paperSize.Width, w.paperSize.Height)
	return true
}

func (w *DefaultWriter) SetFont(family string, style string, size float64) {
	w.impl.SetFont(family, style, size)
}

func (w *DefaultWriter) Text(x float64, y float64, text string) {
	w.impl.Text(x, y, text)
}

func (w *DefaultWriter) WriteTo(writer io.Writer) (int64, error) {
	cw := rw.NewCountWriter(writer)
	err := w.impl.Output(cw)
	return cw.BytesWritten(), err
}

func (w *DefaultWriter) WriteToFile(filepath string) error {
	pdfBytes, err := w.ProduceBytes()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, pdfBytes, 0644)
}

func (w *DefaultWriter) ProduceBytes() ([]byte, error) {
	var buf bytes.Buffer
	err := w.impl.Output(&buf)
	return buf.Bytes(), err
}
