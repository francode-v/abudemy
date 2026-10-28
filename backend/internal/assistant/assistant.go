// Package assistant provides the virtual assistance knowledge base and an
// optional external provider. The built-in knowledge base is a curated set of
// answers written by the platform; it is explicitly not an AI model and never
// replaces medical or emergency professionals. When a provider is configured
// the answers come from it; otherwise the knowledge base is used.
package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"
)

// ErrNotConfigured is returned by a provider that is not available.
var ErrNotConfigured = errors.New("assistant provider not configured")

// Topic is one curated entry of the knowledge base.
type Topic struct {
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Keywords []string `json:"keywords,omitempty"`
	Answer   string   `json:"answer"`
	Related  []string `json:"related,omitempty"`
}

// Answer is the response returned to clients.
type Answer struct {
	Question string  `json:"question,omitempty"`
	Answer   string  `json:"answer"`
	Topic    string  `json:"topic,omitempty"`
	Source   string  `json:"source"`
	AI       bool    `json:"ai"`
	Related  []Topic `json:"related,omitempty"`
}

// KnowledgeBase is the curated, non-AI knowledge base. Answers avoid medical
// advice and point to official channels where appropriate.
var KnowledgeBase = []Topic{
	{
		Slug:     "primeros-pasos",
		Title:    "Cómo crear mi cuenta e iniciar sesión",
		Keywords: []string{"cuenta", "registro", "registrarme", "crear", "iniciar", "sesion", "contraseña", "clave"},
		Answer:   "Para crear tu cuenta toca «Crear cuenta» en la pantalla de inicio, escribe tu nombre, correo y una contraseña de al menos 8 caracteres. Luego inicia sesión con tu correo y contraseña. Si olvidaste tu contraseña, usa «Olvidé mi contraseña» e ingresa tu correo: si está registrado, recibirás instrucciones para restablecerla.",
	},
	{
		Slug:     "asistencia",
		Title:    "Cómo pedir ayuda a mis contactos de confianza",
		Keywords: []string{"ayuda", "asistencia", "emergencia", "contacto", "confianza", "auxilio", "sos"},
		Answer:   "Entra a «Asistencia» y toca «Pedir ayuda». Puedes compartir tu ubicación si lo autorizas. La app avisa a tus contactos de confianza registrados. Importante: la app no es un servicio de emergencia profesional; ante una urgencia vital llama al 131 (SAMU) o 133 (Carabineros).",
	},
	{
		Slug:     "servicios",
		Title:    "Cómo encontrar servicios de salud y farmacias",
		Keywords: []string{"servicio", "salud", "farmacia", "centro", "comunitario", "clinica", "consultorio", "cercano", "mapa"},
		Answer:   "En la pestaña «Servicios» puedes buscar por nombre, filtrar por categoría y comuna, y activar «Cerca de mí» para ordenar por distancia. Toca un servicio para ver su dirección, teléfono, horario y descripción, y desde ahí puedes llamar o abrir la ruta.",
	},
	{
		Slug:     "actividades",
		Title:    "Cómo inscribirme en una actividad o taller",
		Keywords: []string{"actividad", "taller", "curso", "evento", "inscribir", "inscripcion", "participar", "cupo", "lista"},
		Answer:   "En «Actividades» filtra por categoría, fecha, comuna o modalidad. Abre una actividad para ver el detalle, quién la organiza y los requisitos, y toca «Inscribirme». Si no hay cupo, quedarás en lista de espera. Puedes revisar tus participaciones en «Mi cuenta».",
	},
	{
		Slug:     "saberes",
		Title:    "Cómo ofrecer una clase o compartir mis conocimientos",
		Keywords: []string{"clase", "ofrecer", "enseñar", "conocimiento", "saber", "taller propio", "oficio", "compartir"},
		Answer:   "Tus conocimientos son valiosos. En «Mi cuenta» registra tus saberes y luego publica una clase o servicio indicando título, descripción, modalidad y valor. Así otras personas podrán contactarte.",
	},
	{
		Slug:     "oportunidades",
		Title:    "Cómo postular a una oportunidad económica",
		Keywords: []string{"trabajo", "empleo", "oportunidad", "postular", "postulacion", "ingreso", "servicio", "laboral"},
		Answer:   "En «Oportunidades» busca por categoría o comuna, revisa los requisitos y toca «Postular» con un mensaje breve. Puedes ver el estado de tus postulaciones en «Mi cuenta». La plataforma no garantiza empleo ni ingresos: solo conecta personas y organizaciones.",
	},
	{
		Slug:     "estafas",
		Title:    "Qué hacer ante una publicación sospechosa",
		Keywords: []string{"estafa", "sospechoso", "fraude", "reportar", "denunciar", "peligro", "robo"},
		Answer:   "Si una publicación parece sospechosa, toca «Reportar» en su detalle y describe el motivo. Nunca entregues datos bancarios, claves ni pagos por adelantado a desconocidos. Si fuiste víctima de un delito, contacta a Carabineros (133) o a la PDI.",
	},
	{
		Slug:     "accesibilidad",
		Title:    "Cómo cambiar el tamaño de letra y el contraste",
		Keywords: []string{"letra", "tamaño", "contraste", "accesibilidad", "leer", "vision", "texto", "grande"},
		Answer:   "En «Mi cuenta» entra a «Preferencias» y ajusta texto grande, alto contraste o reducción de movimiento. También puedes activar tus intereses para recibir recomendaciones más útiles.",
	},
	{
		Slug:     "privacidad",
		Title:    "Cómo cuido mis datos personales",
		Keywords: []string{"privacidad", "datos", "seguridad", "ubicacion", "permiso", "contraseña"},
		Answer:   "Tu ubicación solo se comparte cuando tú lo autorizas y lo indicas al pedir ayuda. Usa una contraseña larga y no la compartas. Puedes revisar y cerrar las sesiones abiertas en «Seguridad».",
	},
	{
		Slug:     "emergencias",
		Title:    "Números de emergencia oficiales",
		Keywords: []string{"emergencia", "ambulancia", "samu", "carabineros", "bomberos", "fono", "numero", "131", "133", "132"},
		Answer:   "Ante una emergencia llama a los canales oficiales: SAMU 131 (ambulancia), Bomberos 132, Carabineros 133, y 1451 para violencia contra la mujer. La app no reemplaza a los servicios de emergencia profesionales.",
	},
}

// Search returns the best matching topic for a free-text question.
func Search(question string) (Topic, bool) {
	norm := normalize(question)
	if norm == "" {
		return Topic{}, false
	}
	words := strings.Fields(norm)
	best := -1
	bestScore := 0
	for i, t := range KnowledgeBase {
		score := 0
		haystack := normalize(t.Title + " " + strings.Join(t.Keywords, " "))
		for _, w := range words {
			if len(w) < 3 {
				continue
			}
			if strings.Contains(haystack, w) {
				score += 2
			}
			if strings.Contains(normalize(t.Slug), w) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = i
		}
	}
	if best < 0 || bestScore == 0 {
		return Topic{}, false
	}
	return KnowledgeBase[best], true
}

// Related returns up to n topics related to a topic.
func Related(t Topic, n int) []Topic {
	out := make([]Topic, 0, n)
	for _, r := range t.Related {
		for _, kb := range KnowledgeBase {
			if kb.Slug == r {
				out = append(out, kb)
				break
			}
		}
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}

// Provider is an optional external assistant backend.
type Provider interface {
	// Name identifies the provider.
	Name() string
	// Ask returns an answer for a question. It returns ErrNotConfigured when
	// the provider is inactive.
	Ask(ctx context.Context, question string) (string, error)
}

// None is the provider used when no external assistant is configured.
type None struct{}

// Name implements Provider.
func (None) Name() string { return "none" }

// Ask implements Provider and always reports ErrNotConfigured.
func (None) Ask(context.Context, string) (string, error) { return "", ErrNotConfigured }

// Webhook forwards questions to an external HTTP endpoint.
type Webhook struct {
	URL    string
	Token  string
	Client *http.Client
}

// Name implements Provider.
func (Webhook) Name() string { return "webhook" }

// Ask implements Provider.
func (w Webhook) Ask(ctx context.Context, question string) (string, error) {
	body, err := json.Marshal(map[string]string{"question": question})
	if err != nil {
		return "", fmt.Errorf("encode question: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build assistant request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if w.Token != "" {
		req.Header.Set("Authorization", "Bearer "+w.Token)
	}
	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("assistant provider: %w", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("assistant provider returned %d", resp.StatusCode)
	}
	var parsed struct {
		Answer string `json:"answer"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return "", fmt.Errorf("decode assistant response: %w", err)
	}
	if parsed.Answer == "" {
		parsed.Answer = parsed.Text
	}
	if strings.TrimSpace(parsed.Answer) == "" {
		return "", errors.New("assistant provider returned an empty answer")
	}
	return parsed.Answer, nil
}

// New builds a provider from configuration values.
func New(provider, url, token string) Provider {
	if provider == "webhook" && strings.TrimSpace(url) != "" {
		return Webhook{URL: url, Token: token}
	}
	return None{}
}

// SortTopics returns the knowledge base ordered by title (stable output).
func SortTopics() []Topic {
	out := make([]Topic, len(KnowledgeBase))
	copy(out, KnowledgeBase)
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

// normalize lowercases and strips accents so keyword matching is forgiving.
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ä', 'â':
			b.WriteRune('a')
		case 'é', 'è', 'ë', 'ê':
			b.WriteRune('e')
		case 'í', 'ì', 'ï', 'î':
			b.WriteRune('i')
		case 'ó', 'ò', 'ö', 'ô':
			b.WriteRune('o')
		case 'ú', 'ù', 'ü', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
				b.WriteRune(r)
			} else {
				b.WriteRune(' ')
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
