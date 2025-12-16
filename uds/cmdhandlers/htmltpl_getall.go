package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type HtmltplGetAll struct {
	AppProvider framework.AppProviderFunc
}

func (h *HtmltplGetAll) GroupName() string {
	return "htmltpl"
}

func (h *HtmltplGetAll) Command() string {
	return "htmltpl-get-all"
}

func (h *HtmltplGetAll) Desc() string {
	return "Print All Templates in the HTML Template Store"
}

func (h *HtmltplGetAll) Usage() string {
	return h.Command()
}

func (h *HtmltplGetAll) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()

	htmlTplStore := appCore.HTMLTemplateStore
	if htmlTplStore == nil {
		return fmt.Errorf("html template store not ready")
	}
	_, _ = fmt.Fprintln(w, "< File Templates >")
	for key, t := range htmlTplStore.FileTemplates {
		// a key is for a template set
		_, _ = fmt.Fprintf(w, "\n________ Template Set: %s ________\n", key)
		// collect all templates inside the set
		all := t.Templates()
		// each internal template
		for _, tmpl := range all {
			_, _ = fmt.Fprintf(w, "\n\t\t[ %s ]\n\n", tmpl.Name())
			if tmpl.Tree != nil && tmpl.Tree.Root != nil {
				_, _ = fmt.Fprintln(w, tmpl.Tree.Root.String())
			} else {
				_, _ = fmt.Fprintln(w, "(no AST)")
			}
			_, _ = fmt.Fprintln(w, " ")
		}
	}
	// ToDo: Derived
	_, _ = fmt.Fprintln(w, "________________________________________________")
	return nil
}
