package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"rsc.io/pdf"
)

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("ERROR: failed to get response from url: %v", err)
	}

	defer resp.Body.Close()

	out, err := os.Create(filepath)

	if err != nil {
		return fmt.Errorf("ERROR: failed to create out by file path: %v", err)
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}

func ExtractPDFText(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("ERROR: failed to open pdf file by path: %v", err)
	}

	var buf bytes.Buffer

	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		content := p.Content()
		for _, t := range content.Text {
			buf.WriteString(t.S)
			buf.WriteString(" ")
		}
	}

	text := strings.TrimSpace(buf.String())
	if len(text) == 0 {
		return "", fmt.Errorf("no text found in PDF")
	}

	return text, nil
}

// ExtractPPTXText extracts text from PPTX files
func ExtractPPTXText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PPTX file: %w", err)
	}
	defer r.Close()

	var allText strings.Builder
	slideNum := 0

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideNum++
			text, err := extractTextFromSlideXML(f)
			if err != nil {
				continue
			}
			if text != "" {
				allText.WriteString(fmt.Sprintf("\n--- Slide %d ---\n", slideNum))
				allText.WriteString(text)
				allText.WriteString("\n")
			}
		}
	}

	result := strings.TrimSpace(allText.String())
	if result == "" {
		return "", fmt.Errorf("no text found in PPTX")
	}

	return result, nil
}

func extractTextFromSlideXML(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	var texts []string
	decoder := xml.NewDecoder(bytes.NewReader(content))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &se); err == nil && text != "" {
					texts = append(texts, text)
				}
			}
		}
	}

	result := strings.Join(texts, " ")
	re := regexp.MustCompile(`\s+`)
	result = re.ReplaceAllString(result, " ")

	return strings.TrimSpace(result), nil
}

// GetFileExtension returns the lowercase file extension
func GetFileExtension(filename string) string {
	parts := strings.Split(strings.ToLower(filename), ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
