// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Layout(title string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/apps/embeddings/templates/layout.templ`, Line: 9, Col: 17}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</title><link href=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css\" rel=\"stylesheet\"><script src=\"https://unpkg.com/htmx.org@1.9.10\"></script><style>\n\t\t\t\t:root {\n\t\t\t\t\t--primary-color: #00c8c8;\n\t\t\t\t\t--secondary-color: #ff6ac1;\n\t\t\t\t\t--accent-color: #ffd700;\n\t\t\t\t\t--dark-color: #1c162b;\n\t\t\t\t\t--text-color: #e0e0e0;\n\t\t\t\t\t--text-shadow: 0 0 5px rgba(0, 200, 200, 0.7);\n\t\t\t\t\t--glow-effect: 0 0 10px rgba(0, 200, 200, 0.7), 0 0 20px rgba(0, 200, 200, 0.5);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\tbody {\n\t\t\t\t\tbackground-color: var(--dark-color);\n\t\t\t\t\tcolor: var(--text-color);\n\t\t\t\t\tfont-family: 'Courier New', monospace;\n\t\t\t\t\tbackground-image: \n\t\t\t\t\t\tradial-gradient(circle at 25% 25%, rgba(255, 106, 193, 0.1) 0%, transparent 50%),\n\t\t\t\t\t\tradial-gradient(circle at 75% 75%, rgba(0, 200, 200, 0.1) 0%, transparent 50%);\n\t\t\t\t\tpadding-bottom: 2rem;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.container {\n\t\t\t\t\tmax-width: 1200px;\n\t\t\t\t\tpadding: 2rem;\n\t\t\t\t\tbackground-color: rgba(28, 22, 43, 0.8);\n\t\t\t\t\tborder: 1px solid var(--primary-color);\n\t\t\t\t\tborder-radius: 10px;\n\t\t\t\t\tbox-shadow: 0 0 15px rgba(0, 200, 200, 0.3);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\theader {\n\t\t\t\t\ttext-align: center;\n\t\t\t\t\tborder-bottom: 2px solid var(--primary-color) !important;\n\t\t\t\t\tmargin-bottom: 2rem !important;\n\t\t\t\t\tpadding-bottom: 1rem !important;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\th1 {\n\t\t\t\t\tcolor: var(--primary-color);\n\t\t\t\t\ttext-shadow: var(--text-shadow);\n\t\t\t\t\tletter-spacing: 2px;\n\t\t\t\t\tfont-weight: bold;\n\t\t\t\t\ttext-transform: uppercase;\n\t\t\t\t\tfont-size: 2.5rem !important;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.form-control {\n\t\t\t\t\tbackground-color: var(--dark-color);\n\t\t\t\t\tcolor: var(--text-color);\n\t\t\t\t\tborder: 1px solid var(--primary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.form-control:focus {\n\t\t\t\t\tbackground-color: var(--dark-color);\n\t\t\t\t\tcolor: var(--text-color);\n\t\t\t\t\tborder-color: var(--secondary-color);\n\t\t\t\t\tbox-shadow: var(--glow-effect);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.form-floating label {\n\t\t\t\t\tcolor: var(--primary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.form-floating>.form-control:focus~label {\n\t\t\t\t\tcolor: var(--secondary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.card {\n\t\t\t\t\tbackground-color: rgba(28, 22, 43, 0.9);\n\t\t\t\t\tborder: 1px solid var(--primary-color);\n\t\t\t\t\tcolor: var(--text-color);\n\t\t\t\t\toverflow: hidden;\n\t\t\t\t\tposition: relative;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.card::before {\n\t\t\t\t\tcontent: '';\n\t\t\t\t\tposition: absolute;\n\t\t\t\t\ttop: 0;\n\t\t\t\t\tleft: -100%;\n\t\t\t\t\twidth: 100%;\n\t\t\t\t\theight: 4px;\n\t\t\t\t\tbackground: linear-gradient(90deg, var(--primary-color), var(--secondary-color), var(--accent-color));\n\t\t\t\t\tanimation: glowBorder 4s linear infinite;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes glowBorder {\n\t\t\t\t\t0% { left: -100%; }\n\t\t\t\t\t100% { left: 100%; }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.card-header {\n\t\t\t\t\tbackground-color: rgba(0, 200, 200, 0.2);\n\t\t\t\t\tcolor: var(--primary-color);\n\t\t\t\t\tfont-weight: bold;\n\t\t\t\t\ttext-transform: uppercase;\n\t\t\t\t\tletter-spacing: 1px;\n\t\t\t\t\ttext-align: center;\n\t\t\t\t\tborder-bottom: 1px solid var(--primary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.card-body {\n\t\t\t\t\tpadding: 1.5rem;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.similarity-result {\n\t\t\t\t\tfont-size: 3rem !important;\n\t\t\t\t\tfont-weight: bold;\n\t\t\t\t\tcolor: var(--accent-color);\n\t\t\t\t\ttext-shadow: 0 0 10px rgba(255, 215, 0, 0.7);\n\t\t\t\t\tfont-family: 'Digital-7', 'Courier New', monospace;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.result-card {\n\t\t\t\t\ttransition: all 0.3s;\n\t\t\t\t\theight: 100%;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.result-card.htmx-swapping {\n\t\t\t\t\topacity: 0.5;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.lead {\n\t\t\t\t\tcolor: var(--text-color);\n\t\t\t\t\ttext-align: center;\n\t\t\t\t\tfont-size: 1.2rem;\n\t\t\t\t\tmargin-bottom: 2rem;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.text-muted {\n\t\t\t\t\tcolor: rgba(224, 224, 224, 0.6) !important;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.btn-primary {\n\t\t\t\t\tbackground-color: var(--primary-color);\n\t\t\t\t\tborder-color: var(--primary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.btn-primary:hover {\n\t\t\t\t\tbackground-color: var(--secondary-color);\n\t\t\t\t\tborder-color: var(--secondary-color);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Special retro CRT effect */\n\t\t\t\t@keyframes scanline {\n\t\t\t\t\t0% { transform: translateY(0); }\n\t\t\t\t\t100% { transform: translateY(100vh); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\tbody::after {\n\t\t\t\t\tcontent: '';\n\t\t\t\t\tposition: fixed;\n\t\t\t\t\ttop: 0;\n\t\t\t\t\tleft: 0;\n\t\t\t\t\twidth: 100%;\n\t\t\t\t\theight: 2px;\n\t\t\t\t\tbackground: rgba(255, 255, 255, 0.1);\n\t\t\t\t\tz-index: 9999;\n\t\t\t\t\tanimation: scanline 8s linear infinite;\n\t\t\t\t\tpointer-events: none;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Vintage monitor effect */\n\t\t\t\t.container {\n\t\t\t\t\tposition: relative;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.container::before {\n\t\t\t\t\tcontent: '';\n\t\t\t\t\tposition: absolute;\n\t\t\t\t\ttop: 0;\n\t\t\t\t\tleft: 0;\n\t\t\t\t\tright: 0;\n\t\t\t\t\tbottom: 0;\n\t\t\t\t\tbackground: \n\t\t\t\t\t\tlinear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.1) 50%), \n\t\t\t\t\t\tlinear-gradient(90deg, rgba(255, 0, 0, 0.03), rgba(0, 255, 0, 0.03), rgba(0, 0, 255, 0.03));\n\t\t\t\t\tbackground-size: 100% 2px, 3px 100%;\n\t\t\t\t\tpointer-events: none;\n\t\t\t\t\tz-index: 10;\n\t\t\t\t\tborder-radius: 10px;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Comp inputs styling */\n\t\t\t\t.comp-input {\n\t\t\t\t\tposition: relative;\n\t\t\t\t\tmargin-bottom: 1rem;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.input-glow {\n\t\t\t\t\tposition: absolute;\n\t\t\t\t\tbottom: 0;\n\t\t\t\t\tleft: 50%;\n\t\t\t\t\twidth: 50%;\n\t\t\t\t\theight: 2px;\n\t\t\t\t\tbackground: var(--primary-color);\n\t\t\t\t\ttransform: translateX(-50%);\n\t\t\t\t\tfilter: blur(1px);\n\t\t\t\t\topacity: 0.7;\n\t\t\t\t\tbox-shadow: 0 0 10px var(--primary-color);\n\t\t\t\t\tanimation: pulsate 2s ease-in-out infinite;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes pulsate {\n\t\t\t\t\t0% { opacity: 0.5; width: 30%; }\n\t\t\t\t\t50% { opacity: 1; width: 70%; }\n\t\t\t\t\t100% { opacity: 0.5; width: 30%; }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Retro loader */\n\t\t\t\t.retro-loader {\n\t\t\t\t\twidth: 40px;\n\t\t\t\t\theight: 40px;\n\t\t\t\t\tmargin: 1rem auto;\n\t\t\t\t\tborder: 3px solid rgba(0, 200, 200, 0.2);\n\t\t\t\t\tborder-top: 3px solid var(--primary-color);\n\t\t\t\t\tborder-radius: 50%;\n\t\t\t\t\tanimation: spin 1.5s linear infinite;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes spin {\n\t\t\t\t\t0% { transform: rotate(0deg); }\n\t\t\t\t\t100% { transform: rotate(360deg); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Result indicator */\n\t\t\t\t.result-indicator {\n\t\t\t\t\twidth: 80px;\n\t\t\t\t\theight: 80px;\n\t\t\t\t\tmargin: 0 auto;\n\t\t\t\t\tbackground: \n\t\t\t\t\t\tradial-gradient(circle at center, var(--accent-color) 0%, transparent 60%),\n\t\t\t\t\t\tconic-gradient(var(--primary-color), var(--secondary-color), var(--primary-color));\n\t\t\t\t\tborder-radius: 50%;\n\t\t\t\t\topacity: 0.8;\n\t\t\t\t\tbox-shadow: 0 0 15px var(--primary-color);\n\t\t\t\t\tanimation: rotate 10s linear infinite, pulse 3s ease-in-out infinite;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes rotate {\n\t\t\t\t\t0% { transform: rotate(0deg); }\n\t\t\t\t\t100% { transform: rotate(360deg); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes pulse {\n\t\t\t\t\t0% { opacity: 0.5; transform: scale(0.8) rotate(0deg); }\n\t\t\t\t\t50% { opacity: 0.9; transform: scale(1.1) rotate(180deg); }\n\t\t\t\t\t100% { opacity: 0.5; transform: scale(0.8) rotate(360deg); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Retro decoration */\n\t\t\t\t.retro-decoration {\n\t\t\t\t\theight: 6px;\n\t\t\t\t\tbackground: linear-gradient(90deg, \n\t\t\t\t\t\ttransparent 0%, \n\t\t\t\t\t\tvar(--primary-color) 20%, \n\t\t\t\t\t\tvar(--secondary-color) 50%, \n\t\t\t\t\t\tvar(--primary-color) 80%, \n\t\t\t\t\t\ttransparent 100%);\n\t\t\t\t\tmargin: 1rem 0;\n\t\t\t\t\tposition: relative;\n\t\t\t\t\tborder-radius: 3px;\n\t\t\t\t\topacity: 0.8;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.retro-decoration::before, \n\t\t\t\t.retro-decoration::after {\n\t\t\t\t\tcontent: '';\n\t\t\t\t\tposition: absolute;\n\t\t\t\t\twidth: 10px;\n\t\t\t\t\theight: 10px;\n\t\t\t\t\tbackground-color: var(--accent-color);\n\t\t\t\t\tborder-radius: 50%;\n\t\t\t\t\ttop: -2px;\n\t\t\t\t\tanimation: float 3s ease-in-out infinite alternate;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.retro-decoration::before {\n\t\t\t\t\tleft: 20%;\n\t\t\t\t\tanimation-delay: 0.5s;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.retro-decoration::after {\n\t\t\t\t\tright: 20%;\n\t\t\t\t\tanimation-delay: 1s;\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t@keyframes float {\n\t\t\t\t\t0% { transform: translateY(0) scale(1); }\n\t\t\t\t\t100% { transform: translateY(-10px) scale(1.2); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t/* Additional psychedelic effects */\n\t\t\t\t@keyframes rainbow {\n\t\t\t\t\t0% { color: var(--primary-color); }\n\t\t\t\t\t33% { color: var(--secondary-color); }\n\t\t\t\t\t66% { color: var(--accent-color); }\n\t\t\t\t\t100% { color: var(--primary-color); }\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t.card-header {\n\t\t\t\t\tanimation: rainbow 8s linear infinite;\n\t\t\t\t}\n\t\t\t</style></head><body><div class=\"container py-4\"><header class=\"pb-3 mb-4 border-bottom\"><h1>Text Similarity Analysis</h1></header>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</div><script src=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js\"></script></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
