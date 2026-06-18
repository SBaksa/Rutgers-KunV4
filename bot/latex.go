package bot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/bwmarrin/discordgo"
)

var latexRegex = regexp.MustCompile(`(?s)\$\$.+?\$\$`)

const latexPreamble = `\usepackage{amsfonts}
\usepackage{amssymb}
\usepackage{mathrsfs}
\usepackage{mathtools}
\usepackage{esint}
\usepackage{xcolor}
\usepackage{cancel}
\usepackage{dsfont}
\usepackage{bm}
\usepackage{nicefrac}
\usepackage{physics}
\usepackage{pgfplots}
\usepackage{tikz-cd}
\usepackage{mhchem}
\usepackage{chemfig}
\pagecolor{white}
\color{black}`

// qlEscape encodes spaces as %20 (not +) so QuickLaTeX doesn't misinterpret + signs in formulas.
func qlEscape(s string) string {
	escaped := strings.ReplaceAll(s, "%", "%25")
	escaped = strings.ReplaceAll(escaped, " ", "%20")
	escaped = strings.ReplaceAll(escaped, "+", "%2B")
	escaped = strings.ReplaceAll(escaped, "\n", "%0A")
	escaped = strings.ReplaceAll(escaped, "\r", "%0D")
	escaped = strings.ReplaceAll(escaped, "{", "%7B")
	escaped = strings.ReplaceAll(escaped, "}", "%7D")
	escaped = strings.ReplaceAll(escaped, "\\", "%5C")
	escaped = strings.ReplaceAll(escaped, "^", "%5E")
	escaped = strings.ReplaceAll(escaped, "*", "%2A")
	escaped = strings.ReplaceAll(escaped, "[", "%5B")
	escaped = strings.ReplaceAll(escaped, "]", "%5D")
	escaped = strings.ReplaceAll(escaped, "=", "%3D")
	escaped = strings.ReplaceAll(escaped, "&", "%26")
	return escaped
}

func renderLatex(formula string) (string, error) {
	wrapped := fmt.Sprintf("\\begin{equation*}%s\\end{equation*}", formula)

	body := fmt.Sprintf(
		"formula=%s&fsize=%s&fcolor=%s&bcolor=%s&mode=0&out=1&remhost=quicklatex.com&preamble=%s&errors=1",
		qlEscape(wrapped),
		qlEscape("75px"),
		qlEscape("000000"),
		qlEscape("ffffff"),
		qlEscape(latexPreamble),
	)
	resp, err := http.Post("https://quicklatex.com/latex3.f", "application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(string(respBody)), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "0" {
		errMsg := "unknown error"
		if len(lines) >= 2 {
			errMsg = strings.Join(lines[1:], "\n")
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	// line format: "URL [errcode] [width] [height]"
	imgURL := strings.Fields(lines[1])[0]
	return imgURL, nil
}

const latexBorderPx = 20

func addBorder(r io.Reader) (io.Reader, error) {
	src, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()+2*latexBorderPx, b.Dy()+2*latexBorderPx))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(latexBorderPx, latexBorderPx, latexBorderPx+b.Dx(), latexBorderPx+b.Dy()), src, b.Min, draw.Over)
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return &buf, nil
}

func handleLatex(s *discordgo.Session, m *discordgo.MessageCreate, log *logger.Logger) {
	matches := latexRegex.FindAllString(m.Content, -1)
	if len(matches) == 0 {
		return
	}

	for i, match := range matches {
		formula := strings.TrimSpace(match[2 : len(match)-2])
		formula = strings.ReplaceAll(formula, "`", "")
		if formula == "" {
			continue
		}

		imgURL, err := renderLatex(formula)
		if err != nil {
			log.Warn("LaTeX render failed", "error", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("There were issues parsing your LaTeX expression: `%s`", err.Error()))
			continue
		}

		imgResp, err := http.Get(imgURL)
		if err != nil {
			log.Error("Failed to fetch LaTeX image", "url", imgURL, "error", err)
			continue
		}
		defer imgResp.Body.Close()

		padded, err := addBorder(imgResp.Body)
		if err != nil {
			log.Warn("Failed to add border to LaTeX image", "error", err)
			continue
		}

		filename := fmt.Sprintf("latex_%d.png", i)
		s.ChannelFileSend(m.ChannelID, filename, padded)
	}
}
