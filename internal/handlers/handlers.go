// Package handlers provides HTTP request handlers for Learning Desktop.
package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/birddigital/htmx-r/components"
	"github.com/google/uuid"
)

//go:embed templates/*
var templates embed.FS

// PageContext provides common data for all pages
type PageContext struct {
	Title        string
	UserName     string
	SessionID    string
	ActiveNav    string
	NotificationCount int
}

// RenderLayout renders the full dashboard layout
func RenderLayout(w http.ResponseWriter, r *http.Request, ctx PageContext, contentHTML template.HTML) error {
	// Build sidebar navigation items
	sidebarItems := []components.NavItemProps{
		{
			Label:      "Chat",
			Icon:       "M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z",
			HXGet:      "/?partial=true",
			HXTarget:   "#main-content",
			StateValue: "chat",
		},
		{
			Label:      "Courses",
			Icon:       "M4 19.5A2.5 2.5 0 0 1 6.5 17H20",
			HXGet:      "/courses?partial=true",
			HXTarget:   "#main-content",
			StateValue: "courses",
		},
		{
			Label:      "Progress",
			Icon:       "M12 20V10M18 20V4M6 20v-6",
			HXGet:      "/progress?partial=true",
			HXTarget:   "#main-content",
			StateValue: "progress",
		},
		{
			Label:      "Settings",
			Icon:       "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M10.5 12.5l-2.5 2.5",
			HXGet:      "/settings?partial=true",
			HXTarget:   "#main-content",
			StateValue: "settings",
		},
	}

	layout := components.Layout(components.LayoutProps{
		Sidebar: components.SidebarProps{
			BrandName:  "Learning Desktop",
			BrandShort: "LD",
			Items:      sidebarItems,
			StateKey:   "sidebar",
		},
		PageTitle:         ctx.Title,
		UserName:          ctx.UserName,
		NotificationCount: ctx.NotificationCount,
		ActiveNav:         ctx.ActiveNav,
		ContentHTML:       contentHTML,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return layout.Render(w)
}

// RenderPartial renders just the page content (for HTMX navigation)
func RenderPartial(w http.ResponseWriter, r *http.Request, contentHTML template.HTML) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(string(contentHTML)))
}

// getOrCreateSessionID gets or creates a session ID from cookie
func GetOrCreateSessionID(r *http.Request) string {
	if cookie, err := r.Cookie("session_id"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return uuid.New().String()
}

// renderTemplate renders an embedded template
func renderTemplate(name string, data interface{}) (template.HTML, error) {
	tmpl, err := template.ParseFS(templates, "templates/"+name+".html")
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}
