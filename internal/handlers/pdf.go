package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
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

    // В rsc.io/pdf страницы нумеруются с 1
    for i := 1; i <= r.NumPage(); i++ {
        p := r.Page(i)
        if p.V.IsNull() { // проверка на пустую страницу
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